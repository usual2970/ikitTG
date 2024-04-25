package hash

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5(str string) string {

	// 创建一个MD5哈希对象
	hash := md5.New()

	// 将字符串转换为字节数组并计算哈希值
	hash.Write([]byte(str))
	hashValue := hash.Sum(nil)

	// 将哈希值转换为16进制字符串并打印输出
	hashString := hex.EncodeToString(hashValue)
	return hashString
}
