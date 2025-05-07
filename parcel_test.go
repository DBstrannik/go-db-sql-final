package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	)`)
	require.NoError(t, err)

	return db
}

func TestAddGetDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавляем посылку
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Получаем добавленную посылку
	storedParcel, err := store.Get(id)
	require.NoError(t, err)

	// Сравниваем структуры целиком
	expectedParcel := parcel
	expectedParcel.Number = id
	require.Equal(t, expectedParcel, storedParcel)

	// Удаляем посылку
	err = store.Delete(id)
	require.NoError(t, err)

	// Проверяем, что посылку больше нельзя получить
	_, err = store.Get(id)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавляем посылку
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Меняем адрес
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// Проверяем изменение адреса
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address)

	// Меняем статус на "sent"
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// Пытаемся снова изменить адрес
	err = store.SetAddress(id, "another address")
	require.ErrorIs(t, err, sql.ErrNoRows)

	// Проверяем, что адрес не изменился
	updatedParcel, err = store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address)
}

func TestSetStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавляем посылку
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Меняем статус на "sent"
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// Проверяем изменение статуса
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, updatedParcel.Status)

	// Меняем статус на "delivered"
	err = store.SetStatus(id, ParcelStatusDelivered)
	require.NoError(t, err)

	// Проверяем изменение статуса
	updatedParcel, err = store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusDelivered, updatedParcel.Status)
}

func TestGetByClient(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := make(map[int]Parcel)

	// Задаём всем посылкам один clientId
	clientId := randRange.Intn(10_000_000)
	parcels[0].Client = clientId
	parcels[1].Client = clientId
	parcels[2].Client = clientId

	// Добавляем посылки
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// Получаем посылки клиента
	storedParcels, err := store.GetByClient(clientId)
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))

	// Проверяем посылки
	for _, parcel := range storedParcels {
		expected, exists := parcelMap[parcel.Number]
		require.True(t, exists)
		require.Equal(t, expected, parcel)
	}
}
