package server

import (
	"context"
	"testing"
)

func TestDBConnection(t *testing.T) {
	ctx := context.Background()
	db, err := NewDB(ctx)
	if err != nil {
		t.Errorf("Connection to the database failed, %v", err)
	}
	defer db.Close()
}
