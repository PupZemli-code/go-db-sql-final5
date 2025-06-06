package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// 123
func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return 0, fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	ros, err := db.Exec("INSERT INTO parcel(client, status, address, created_at) VALUES(:client, :status, :address, :created_at)",
		//sql.Named("number", p.Number),
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}
	id, err := ros.LastInsertId()
	if err != nil {
		return 0, err
	}
	// верните идентификатор последней добавленной записи
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	p := Parcel{}
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return p, fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	// заполните объект Parcel данными из таблицы
	row := db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = :number", sql.Named("number", number))
	err = row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return p, err
	}
	if err != nil {
		return p, fmt.Errorf("ошибка чтения: %v", err)
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк

	// заполните срез Parcel данными из таблицы
	var res []Parcel
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return res, fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", client))
	if err == sql.ErrNoRows {
		return res, fmt.Errorf("в базе данных нет строки: %v", err)
	}
	if err != nil {
		return res, fmt.Errorf("ошибка чтения: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		p := Parcel{}

		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return res, fmt.Errorf("ошибка чтения: %v", err)
		}
		res = append(res, p)
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	_, err = db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("number", number),
		sql.Named("status", status))
	if err == sql.ErrNoRows {
		return fmt.Errorf("в базе данных нет строки: %v", err)
	}
	if err != nil {
		return fmt.Errorf("ошибка чтения: %v", err)
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	var status string
	row := db.QueryRow("SELECT status FROM parcel WHERE number = :number",
		sql.Named("number", number))
	err = row.Scan(&status)
	if err != nil {
		return fmt.Errorf("ошибка чтения: %v", err)
	}
	if status != ParcelStatusRegistered {
		return fmt.Errorf("статус посылки [%s], для изменения адреса, статус должен быть [%s]", status, ParcelStatusRegistered)
	}

	_, err = db.Exec("UPDATE parcel SET address = :address WHERE number = :number",
		sql.Named("number", number),
		sql.Named("address", address))
	if err == sql.ErrNoRows {
		return fmt.Errorf("в базе данных нет строки: %v", err)
	}
	if err != nil {
		return fmt.Errorf("ошибка чтения: %v", err)
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	var status string
	row := db.QueryRow("SELECT status FROM parcel WHERE number = :number",
		sql.Named("number", number))
	err = row.Scan(&status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Строка не найдена")
	}
	if err != nil {
		return fmt.Errorf("Delete row.Scan ошибка чтения: %v", err)
	}
	if status != ParcelStatusRegistered {
		return fmt.Errorf("статус посылки [%s], для удаления, статус должен быть [%s]", status, ParcelStatusRegistered)
	}

	_, err = db.Exec("DELETE FROM parcel WHERE number = :number", sql.Named("number", number))
	if err == sql.ErrNoRows {
		return fmt.Errorf("Строка не найдена")
	}
	return nil
}
func (s ParcelStore) MasterDelete(number int) error {
	// реализуйте удаление строки из таблицы parcel не зависимо от статуса
	// удалять строку можно только если значение статуса registered
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	var status string
	row := db.QueryRow("SELECT status FROM parcel WHERE number = :number",
		sql.Named("number", number))
	err = row.Scan(&status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Строка не найдена")
	}
	if err != nil {
		return fmt.Errorf("Delete row.Scan ошибка чтения: %v", err)
	}

	_, err = db.Exec("DELETE FROM parcel WHERE number = :number", sql.Named("number", number))
	if err == sql.ErrNoRows {
		return fmt.Errorf("Строка не найдена")
	}
	return nil
}
