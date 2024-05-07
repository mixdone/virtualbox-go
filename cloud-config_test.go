package virtualbox

import (
	"errors"
	"testing"
)

type CmdN struct {
	stdout string
	stderr string
	err    error
}

func (fc *CmdN) setOptions(opts ...option) Command {
	return fc
}

func (fc *CmdN) isGuest() bool {
	return false
}

func (fc *CmdN) path() string {
	return "fake"
}

func (fc *CmdN) run(args ...string) (string, string, error) {
	// Эмулируем вызов VBoxManage setextradata и getextradata
	if len(args) >= 3 && args[0] == "setextradata" {
		// Проверяем, что команда setextradata была вызвана с правильными аргументами
		if args[1] == "TestVM" && args[2] == "testKey" {
			return "", "", nil
		}
	} else if len(args) >= 3 && args[0] == "getextradata" {
		// Проверяем, что команда getextradata была вызвана с правильными аргументами
		if args[1] == "TestVM" && args[2] == "testKey" {
			return "Value: testValue", "", nil
		}
	}

	return "", "", errors.New("unsupported command")
}

func TestSetAndGetCloudData(t *testing.T) {
	vbox := &VBox{Name: "TestVM"}

	// Замена manage
	manage = &CmdN{
		stdout: "",
		stderr: "",
		err:    nil,
	}
	defer func() {
		manage = nil // Возвращаем manage обратно к исходному состоянию
	}()

	key := "testKey"
	value := "testValue"
	err := vbox.SetCloudData(key, value)
	if err != nil {
		t.Fatalf("Failed to set test data: %v", err)
	}

	retrievedValue, err := vbox.GetCloudData(key)

	if err != nil {
		t.Errorf("GetCloudData failed: %v", err)
	}

	// Проверяем полученное значение
	expectedValue := "testValue"
	if retrievedValue == nil || *retrievedValue != expectedValue {
		t.Errorf("Retrieved value does not match expected value: expected %q, got %q", expectedValue, *retrievedValue)
	}
}
