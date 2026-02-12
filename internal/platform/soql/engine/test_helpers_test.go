package engine

import (
	"context"
	"sync"
)

// testCache is a simple in-memory cache for testing purposes.
type testCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
	stats CacheStats
}

func newTestCache[K comparable, V any]() *testCache[K, V] {
	return &testCache[K, V]{
		items: make(map[K]V),
	}
}

func (c *testCache[K, V]) Get(_ context.Context, key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, found := c.items[key]
	if found {
		c.stats.Hits++
	} else {
		c.stats.Misses++
	}
	return val, found
}

func (c *testCache[K, V]) Set(_ context.Context, key K, value V) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
	c.stats.Size = int64(len(c.items))
	return nil
}

func (c *testCache[K, V]) Delete(_ context.Context, key K) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	c.stats.Size = int64(len(c.items))
	return nil
}

func (c *testCache[K, V]) Clear(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]V)
	c.stats.Size = 0
	return nil
}

func (c *testCache[K, V]) Remove(predicate func(V) bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if predicate(v) {
			delete(c.items, k)
		}
	}
	c.stats.Size = int64(len(c.items))
	return nil
}

func (c *testCache[K, V]) GetStats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return &CacheStats{
		Hits:   c.stats.Hits,
		Misses: c.stats.Misses,
		Size:   int64(len(c.items)),
	}
}

// setupTestMetadata creates a MetadataProvider with test data.
func setupTestMetadata() MetadataProvider {
	objects := map[string]*ObjectMeta{
		"Account": NewObjectMeta("Account", "", "accounts").
			Field("Id", "id", FieldTypeID).
			Field("Name", "name", FieldTypeString).
			Field("Industry", "industry", FieldTypeString).
			Field("AnnualRevenue", "annual_revenue", FieldTypeFloat).
			Field("CreatedDate", "created_at", FieldTypeDateTime).
			Field("OwnerId", "owner_id", FieldTypeID).
			Lookup("Owner", "owner_id", "User", "id").
			Relationship("Contacts", "Contact", "account_id", "id").
			Relationship("Opportunities", "Opportunity", "account_id", "id").
			Build(),

		"Contact": NewObjectMeta("Contact", "", "contacts").
			Field("Id", "id", FieldTypeID).
			Field("Name", "name", FieldTypeString).
			Field("FirstName", "first_name", FieldTypeString).
			Field("LastName", "last_name", FieldTypeString).
			Field("Email", "email", FieldTypeString).
			Field("Phone", "phone", FieldTypeString).
			Field("AccountId", "account_id", FieldTypeID).
			Field("CreatedDate", "created_at", FieldTypeDateTime).
			Lookup("Account", "account_id", "Account", "id").
			Build(),

		"Opportunity": NewObjectMeta("Opportunity", "", "opportunities").
			Field("Id", "id", FieldTypeID).
			Field("Name", "name", FieldTypeString).
			Field("Amount", "amount", FieldTypeFloat).
			Field("StageName", "stage_name", FieldTypeString).
			Field("CloseDate", "close_date", FieldTypeDate).
			Field("AccountId", "account_id", FieldTypeID).
			Field("OwnerId", "owner_id", FieldTypeID).
			Lookup("Account", "account_id", "Account", "id").
			Lookup("Owner", "owner_id", "User", "id").
			Build(),

		"User": NewObjectMeta("User", "", "users").
			Field("Id", "id", FieldTypeID).
			Field("Name", "name", FieldTypeString).
			Field("Email", "email", FieldTypeString).
			Field("ManagerId", "manager_id", FieldTypeID).
			Lookup("Manager", "manager_id", "User", "id").
			Build(),

		"Task": NewObjectMeta("Task", "", "tasks").
			Field("Id", "id", FieldTypeID).
			Field("Subject", "subject", FieldTypeString).
			Field("Description", "description", FieldTypeString).
			Field("Status", "status", FieldTypeString).
			Field("WhatId", "what_id", FieldTypeID). // Polymorphic field for Account/Opportunity
			Field("WhoId", "who_id", FieldTypeID).   // Polymorphic field for Contact/Lead
			Field("OwnerId", "owner_id", FieldTypeID).
			Field("CreatedDate", "created_at", FieldTypeDateTime).
			Lookup("Owner", "owner_id", "User", "id").
			Build(),

		"Lead": NewObjectMeta("Lead", "", "leads").
			Field("Id", "id", FieldTypeID).
			Field("Name", "name", FieldTypeString).
			Field("Email", "email", FieldTypeString).
			Field("Company", "company", FieldTypeString).
			Field("Status", "status", FieldTypeString).
			Build(),
	}

	return NewStaticMetadataProvider(objects)
}
