package mess

import (
	"testing"
	"time"
)

func TestA(t *testing.T) {
	t.Cleanup(func() {})
}

func TestB(t *testing.T) {
	t.Error("failed")
}

func TestC(t *testing.T) {
	t.Parallel()
	time.Sleep(time.Second)
	t.Fatal("failed")
}

func TestD(t *testing.T) {
	t.Parallel()
	t.Fail()
}

func TestE(t *testing.T) {
	t.Skip()
	t.Fatalf("failed %v", "TestE")
}

func TestF(t *testing.T) {
	t.Run("testg", func(t *testing.T) {
	})
	t.Run("testh", func(t *testing.T) {
		t.Parallel()
		t.Fatal("failed")
	})
	t.Run("testi", func(t *testing.T) {
		t.Skipf("Skipped with val: %v", "clippy")
	})
}

func TestJ(t *testing.T) {
	t.Errorf("This is an error with val: %v", 42)
}

func TestK(t *testing.T) {
	t.Log("Logging this test")
	t.Logf("Logging this value: %v", "wow")
	t.Fatal("failed")
}
