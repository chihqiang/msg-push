package sender

import (
	"testing"
)

func TestProviderMetasAll(t *testing.T) {
	metas := GetAll()
	// 8 个内置服务商
	if len(metas) != 8 {
		t.Fatalf("GetAll len = %d, want 8", len(metas))
	}
	// 返回的是拷贝：修改不应污染内部注册表
	metas[0].Name = "changed"
	if providerMetas[0].Name == "changed" {
		t.Error("GetAll returned shared Meta reference")
	}
}

func TestProviderMetasByType(t *testing.T) {
	sms := GetByType(TypeSMS)
	if len(sms) != 3 {
		t.Errorf("TypeSMS metas len = %d, want 3", len(sms))
	}
	for _, m := range sms {
		if m.Type != TypeSMS {
			t.Errorf("meta %s type = %s, want sms", m.Code, m.Type)
		}
	}
}

func TestProviderMetasByCode(t *testing.T) {
	m, ok := GetByCode(CodeAliyunSMS)
	if !ok || m == nil {
		t.Fatal("GetByCode(aliyun_sms) should exist")
	}
	if m.Code != CodeAliyunSMS || m.Name != "阿里云短信" {
		t.Errorf("GetByCode meta = %+v", m)
	}
	if _, ok := GetByCode("nonexistent"); ok {
		t.Error("GetByCode(nonexistent) should be false")
	}
}
