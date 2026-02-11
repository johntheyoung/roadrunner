package beeperapi

import (
	"testing"

	beeperdesktopapi "github.com/beeper/desktop-api-go"
)

func TestAccountNetworkMapNil(t *testing.T) {
	if got := accountNetworkMap(nil); got != nil {
		t.Fatalf("accountNetworkMap(nil) = %#v, want nil", got)
	}
}

func TestAccountNetworkMapFiltersInvalidEntries(t *testing.T) {
	accounts := []beeperdesktopapi.Account{
		{AccountID: "acc-whatsapp", Network: "WhatsApp"},
		{AccountID: "", Network: "Telegram"},
		{AccountID: "acc-empty-network", Network: ""},
		{AccountID: "acc-telegram", Network: "Telegram"},
	}

	got := accountNetworkMap(&accounts)
	if got == nil {
		t.Fatal("accountNetworkMap returned nil, want populated map")
	}

	if got["acc-whatsapp"] != "WhatsApp" {
		t.Fatalf("network for acc-whatsapp = %q, want %q", got["acc-whatsapp"], "WhatsApp")
	}
	if got["acc-telegram"] != "Telegram" {
		t.Fatalf("network for acc-telegram = %q, want %q", got["acc-telegram"], "Telegram")
	}
	if _, ok := got["acc-empty-network"]; ok {
		t.Fatalf("unexpected mapping for acc-empty-network: %q", got["acc-empty-network"])
	}
	if _, ok := got[""]; ok {
		t.Fatalf("unexpected mapping for empty account ID: %q", got[""])
	}
}
