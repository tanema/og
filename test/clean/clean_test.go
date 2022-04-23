package clean

import (
	"testing"
	"time"
)

func TestA(t *testing.T) {
}

func TestB(t *testing.T) {
}

func TestC(t *testing.T) {
	t.Parallel()
	time.Sleep(time.Second)
}

func TestD(t *testing.T) {
	t.Parallel()
}

func TestE(t *testing.T) {
	t.Parallel()
}

func TestF(t *testing.T) {
	t.Run("testg", func(t *testing.T) {
	})
	t.Run("testh", func(t *testing.T) {
		t.Parallel()
	})
	t.Run("testi", func(t *testing.T) {
	})
}
