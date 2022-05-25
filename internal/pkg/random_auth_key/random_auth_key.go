package random_auth_key

import (
	"encoding/base64"
	"fmt"

	"github.com/jinzhu/now"
	"github.com/wumansgy/goEncrypt"
)

// RandomAuthKey 伪随机的登录验证 Key
type RandomAuthKey struct {
	offset  int
	authKey AuthKey // 基础的密钥，密钥会基于这个基础的密钥生成
}

// NewRandomAuthKey offset 建议是 5
func NewRandomAuthKey(offset int, authKey AuthKey) *RandomAuthKey {

	tmpOffset := offset
	if tmpOffset < 0 || tmpOffset > 5 {
		tmpOffset = 5
	}
	return &RandomAuthKey{
		offset:  tmpOffset,
		authKey: authKey,
	}
}

// GetAuthKey 获取这个小时的认证密码
func (r *RandomAuthKey) GetAuthKey() (string, error) {

	// 当前时间 Unix时间戳 1653099199，这里获取到的是整点的小时时间
	nowUnixTime := now.BeginningOfHour().Unix()

	return r.getAuthKey(nowUnixTime)
}

func (r RandomAuthKey) getAuthKey(hourUnixTime int64) (string, error) {

	nowUnixTimeStr := fmt.Sprintf("%d", hourUnixTime)
	prefixStr := nowUnixTimeStr[len(nowUnixTimeStr)-r.offset:] + nowUnixTimeStr[:r.offset]
	// 拼接
	orgString := prefixStr + r.authKey.BaseKey + nowUnixTimeStr + r.authKey.BaseKey[:r.offset]
	plaintext := []byte(orgString)
	// 传入明文和自己定义的密钥，密钥为 16 字节 可以自己传入初始化向量,如果不传就使用默认的初始化向量, 16 字节
	cryptText, err := goEncrypt.AesCbcEncrypt(plaintext, []byte(r.authKey.AESKey16), []byte(r.authKey.AESIv16))
	if err != nil {
		return "", err
	}
	// 加密后的字符串
	return base64.StdEncoding.EncodeToString(cryptText), nil

}

type AuthKey struct {
	BaseKey  string // 基础的密钥，密钥会基于这个基础的密钥生成
	AESKey16 string // AES密钥
	AESIv16  string // 初始化向量
}

const (
	BaseKey  = "0123456789123456789" // 基础的密钥，密钥会基于这个基础的密钥生成
	AESKey16 = "1234567890123456"    // AES密钥
	AESIv16  = "1234567890123456"    // 初始化向量
)
