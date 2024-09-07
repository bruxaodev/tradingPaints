package tradingPaints

import (
	"testing"
)

func TestUpdate(t *testing.T) {
	users := []Player{
		{UserId: "337083", CarName: "ferrari296gt3"},
	}
	err := Init()
	if err != nil {
		t.Errorf("Error: %v\n", err)
		return
	}
	err = Update(users, true)
	if err != nil {
		t.Errorf("Error on update: %v", err)
	}
}
