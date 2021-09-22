package proxy_helper

import "testing"

func TestProxyTest(t *testing.T) {

	gotSpeed, gotStatus, err := ProxyTest("http://192.168.50.252:20172")
	if err != nil {
		t.Fatal(err)
	}
	println("Speed:", gotSpeed, "Status:", gotStatus)

	_, _, err = ProxyTest("http:/192.168.1.123:123")
	if err == nil {
		t.Fatal(err)
	}

	_, _, err = ProxyTest("http://192.168.1.123:123")
	if err == nil {
		t.Fatal(err)
	}
}
