package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки

func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Printf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	//db.Exec("DELETE FROM parcel")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	assert.NoError(t, err, "Add ошибка != nil", err)
	assert.NotEmpty(t, id)
	parcel.Number = id
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	p, err := store.Get(parcel.Number)
	require.NoError(t, err, "Get ошибка != nil")
	assert.Equal(t, p.Number, parcel.Number)
	assert.Equal(t, p.Client, parcel.Client)
	assert.Equal(t, p.Address, parcel.Address)
	assert.Equal(t, p.CreatedAt, parcel.CreatedAt)
	assert.Equal(t, p.Status, parcel.Status)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(parcel.Number)
	require.NoError(t, err, "Delete ошибка != nil")
	err = store.Delete(parcel.Number)
	assert.Equal(t, err, fmt.Errorf("cтрока не найдена: sql: no rows in result set"))
	// удоляет добавленную строку из таблицы
	_, err = store.Get(id)
	require.EqualError(t, err, "sql: no rows in result set")
}

// TestSetAddress проверяет обновление адреса

func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Printf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	assert.NoError(t, err, "Add ошибка != nil", err)
	assert.NotEmpty(t, id, "ID отсутствует")
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	assert.NoError(t, err, "SetAddress ошибка != nil", err)
	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	parcel, err = store.Get(id)
	assert.NoError(t, err, "Get ошибка != nil", err)
	assert.Equal(t, newAddress, parcel.Address)

	err = store.Delete(id)
	assert.NoError(t, err, "Delete ошибка != nil", err)
}

// TestSetStatus проверяет обновление статуса

func TestSetStatus(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Printf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	assert.NoError(t, err, "Add ошибка != nil", err)
	assert.NotEmpty(t, id, "ID отсутствует")

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(id, ParcelStatusSent)
	assert.NoError(t, err, "SetStatus ошибка != nil", err)
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	parcel, err = store.Get(id)
	require.NoError(t, err, "Get ошибка != nil", err)
	assert.Equal(t, parcel.Status, ParcelStatusSent)
	err = store.MasterDelete(id)
	require.NoError(t, err)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента

func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Printf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		parcel := parcels[i]
		id, err := store.Add(parcel)
		assert.NoError(t, err, "Add ошибка != nil", err)
		assert.NotEmpty(t, id, "ID отсутствует")

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	require.NoError(t, err, "GetByClient ошибка != nil", err)
	assert.Len(t, storedParcels, 3, "количество полученных посылок не совпадает с количеством добавленных")
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, parcel, parcelMap[parcel.Number])
	}
}
