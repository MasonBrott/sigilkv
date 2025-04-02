package main

import (
	"testing"
)

func TestPut(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		value   string
		initial map[string]string // State of the store *before* this test case's Put
	}{
		{
			name:    "simple put",
			key:     "key1",
			value:   "value1",
			initial: map[string]string{},
		},
		{
			name:    "overwrite existing key",
			key:     "key1",
			value:   "newValue",
			initial: map[string]string{"key1": "oldValue"}, // Pre-populate for overwrite
		},
		{
			name:    "put with empty key",
			key:     "",
			value:   "valueForEmptyKey",
			initial: map[string]string{},
		},
		{
			name:    "put with empty value",
			key:     "keyForEmptyValue",
			value:   "",
			initial: map[string]string{},
		},
		{
			name:    "put with empty key and value",
			key:     "",
			value:   "",
			initial: map[string]string{},
		},
	}

	for _, testcase := range testCases {
		// Use t.Run to create a subtest for each case
		t.Run(testcase.name, func(t *testing.T) {
			// Reset the global store and populate initial state for this specific test case
			store = make(map[string]string)
			for k, v := range testcase.initial {
				store[k] = v
			}

			// Use t.Cleanup to ensure the store is reset after this subtest finishes
			t.Cleanup(func() {
				store = make(map[string]string) // Reset for the next test case
			})

			err := Put(testcase.key, testcase.value)
			if err != nil {
				t.Fatalf("Put(%q, %q) returned an unexpected error: %v", testcase.key, testcase.value, err)
			}

			// Verify the value was stored correctly
			storedValue, ok := store[testcase.key]
			if !ok {
				t.Errorf("Value for key %q not found in store after Put", testcase.key)
			} else if storedValue != testcase.value {
				t.Errorf("Value stored for key %q was incorrect: got %q, want %q", testcase.key, storedValue, testcase.value)
			}

			// Optional: Check if other keys were unexpectedly added/modified (more relevant in complex cases)
			if len(store) > len(testcase.initial) && testcase.initial[testcase.key] == "" { // Check only if it was a new key
				if len(store) != len(testcase.initial)+1 {
					t.Errorf("Store size is incorrect after Put: got %d, expected %d or %d", len(store), len(testcase.initial), len(testcase.initial)+1)
				}
			} else if len(store) != len(testcase.initial) && testcase.initial[testcase.key] != "" { // Check if it was an overwrite
				if len(store) != len(testcase.initial) {
					t.Errorf("Store size is incorrect after Put (overwrite): got %d, expected %d", len(store), len(testcase.initial))
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name          string
		key           string
		expectedValue string
		initial       map[string]string
		expectedErr   error // Added to specify the expected error
	}{
		{
			name:          "simple get",
			key:           "key1",
			expectedValue: "val1",                            // Expected value when key exists
			initial:       map[string]string{"key1": "val1"}, // Must pre-populate the store
			expectedErr:   nil,                               // No error expected
		},
		{
			name:          "get non-existent key",
			key:           "non-existent",
			expectedValue: "",                                // Expect zero value for string on error
			initial:       map[string]string{"key1": "val1"}, // Can have other keys
			expectedErr:   ErrorNoSuchKey,                    // Expect this specific error
		},
		{
			name:          "get existing empty key", // Test retrieving an existing key that is ""
			key:           "",
			expectedValue: "valueForEmptyKey",
			initial:       map[string]string{"": "valueForEmptyKey"},
			expectedErr:   nil,
		},
		{
			name:          "get non-existent empty key", // Test getting "" when it doesn't exist
			key:           "",
			expectedValue: "",
			initial:       map[string]string{"key1": "val1"},
			expectedErr:   ErrorNoSuchKey,
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			store = make(map[string]string)
			for k, v := range testcase.initial {
				store[k] = v
			}

			t.Cleanup(func() {
				store = make(map[string]string)
			})

			value, err := Get(testcase.key)

			// Check if the error matches the expected error
			if err != testcase.expectedErr {
				// Specific check for ErrorNoSuchKey, as errors.Is might be better for wrapped errors
				if !(err == ErrorNoSuchKey && testcase.expectedErr == ErrorNoSuchKey) {
					t.Fatalf("Get(%q): unexpected error: got %v, want %v", testcase.key, err, testcase.expectedErr)
				}
			}

			// Only check the value if the error was as expected (either nil or the correct error type)
			if err == testcase.expectedErr && value != testcase.expectedValue {
				t.Errorf("Get(%q): value incorrect: got %q, want %q", testcase.key, value, testcase.expectedValue)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		initial map[string]string
	}{
		{
			name:    "simple delete",
			key:     "key1",
			initial: map[string]string{"key1": "val1"},
		},
		{
			name:    "delete non-existent key",
			key:     "non-existent",
			initial: map[string]string{"key1": "val1"},
		},
		{
			name:    "delete existing empty key",
			key:     "",
			initial: map[string]string{"": "valueForEmptyKey"},
		},
		{
			name:    "delete non-existent empty key",
			key:     "",
			initial: map[string]string{"key1": "val1"},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			store = make(map[string]string)
			for k, v := range testcase.initial {
				store[k] = v
			}

			t.Cleanup(func() {
				store = make(map[string]string)
			})

			// Store initial state for size comparison later
			initialSize := len(store)
			_, keyExistedInitially := store[testcase.key]

			err := Delete(testcase.key)
			if err != nil {
				t.Fatalf("Delete(%q): unexpected error: got %v, want nil", testcase.key, err)
			}

			// Verify the key is actually gone
			_, keyExistsAfterDelete := store[testcase.key]
			if keyExistsAfterDelete {
				t.Errorf("key %q was not deleted from store", testcase.key)
			}

			// Verify the map size changed appropriately
			finalSize := len(store)
			expectedSize := initialSize
			if keyExistedInitially {
				expectedSize = initialSize - 1
			}

			if finalSize != expectedSize {
				t.Errorf("Store size incorrect after deleting key %q: got %d, want %d (key existed initially: %t)", testcase.key, finalSize, expectedSize, keyExistedInitially)
			}
		})
	}
}
