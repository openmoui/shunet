package rsa

import (
	"shunet/bigint"
	"strings"
)

type RSAPair struct {
	E         bigint.BigInt
	D         bigint.BigInt
	M         bigint.BigInt
	ChunkSize int
	Radix     int
	Barrett   barrettMu
}

func NewRSAPair(encryptionExponent, decryptionExponent, modulus string) *RSAPair {
	res := &RSAPair{
		E:     BiFromHex(encryptionExponent),
		D:     BiFromHex(decryptionExponent),
		M:     BiFromHex(modulus),
		Radix: 16,
	}
	res.ChunkSize = 2 * res.M.BiHighIndex()
	res.Barrett = newBarrettMu(res.M)
	return res
}

func charToHex(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c) - '0'
	case c >= 'A' && c <= 'Z':
		return int(c) - 'A' + 10
	case c >= 'a' && c <= 'z':
		return int(c) - 'a' + 10
	default:
		return 0
	}
}

func hexToDigit(s string) int {
	result := 0
	sl := len(s)
	if sl > 4 {
		sl = 4
	}

	for i := 0; i < sl; i++ {
		result <<= 4
		result |= charToHex(s[i])
	}
	return result
}

func BiFromHex(s string) bigint.BigInt {
	sl := len(s)
	result := bigint.NewBigInt()

	for i, j := sl, 0; i > 0; i, j = i-4, j+1 {
		start := max(0, i-4)
		substr := s[start:i]
		result.Digits[j] = hexToDigit(substr)
	}
	return result
}

var hexatrigesimalToChar = []byte{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
	'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
	'u', 'v', 'w', 'x', 'y', 'z',
}

// biToString 将 BigInt 转换为指定进制的字符串表示
func biToString(x bigint.BigInt, radix int) string {
	b := bigint.NewBigInt()
	b.Digits[0] = radix
	q, r := bigint.BiDivideModulo(x, b)
	result := string(hexatrigesimalToChar[r.Digits[0]])

	for bigint.BiCompare(q, bigint.BigZero) == 1 {
		q, r = bigint.BiDivideModulo(q, b)
		result += string(hexatrigesimalToChar[r.Digits[0]])
	}

	if x.IsNeg {
		result = "-" + result
	}

	return reverseStr(result)
}

func reverseStr(s string) string {
	bytes := []byte(s)
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return string(bytes)
}

var hexToChar = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'a', 'b', 'c', 'd', 'e', 'f'}

// digitToHex 将一个整数转换为十六进制字符串
func digitToHex(n int) string {
	mask := 0xf
	result := ""
	for i := 0; i < 4; i++ {
		result += string(hexToChar[n&mask])
		n >>= 4 // 在 Go 中使用 `>>` 右移位运算符
	}
	return reverseStr(result)
}

// biToHex 将 BigInt 转换为十六进制字符串
func biToHex(x bigint.BigInt) string {
	result := ""
	n := x.BiHighIndex()
	for i := n; i >= 0; i-- {
		result += digitToHex(x.Digits[i])
	}
	return result
}

func (rsa *RSAPair) Encrypt(s string) string {
	a := make([]int, len(s))
	sl := len(s)

	// 将字符串转换为字符码点数组
	for i := 0; i < sl; i++ {
		a[i] = int(s[i])
	}

	// 填充到块大小的倍数
	for len(a)%rsa.ChunkSize != 0 {
		a = append(a, 0)
	}

	al := len(a)
	var result strings.Builder

	for i := 0; i < al; i += rsa.ChunkSize {
		block := bigint.NewBigInt()
		for j, k := 0, i; k < i+rsa.ChunkSize; j++ {
			block.Digits[j] = a[k]
			if k+1 < al {
				block.Digits[j] += a[k+1] << 8
			}
			k += 2
		}

		// 执行模幂运算
		crypt := rsa.Barrett.BarrettMu_powMod(block, rsa.E)

		// 根据 radix 转换为字符串
		var text string
		if rsa.Radix == 16 {
			text = biToHex(crypt)
		} else {
			text = biToString(crypt, rsa.Radix)
		}

		result.WriteString(text + " ")
	}

	// 返回结果字符串，去掉最后的空格
	return strings.TrimSpace(result.String())
}

func (rsa *RSAPair) EncryptedPassword(password, mac string) string {
	passwordMac := password + ">" + mac
	var passwordEncode = reverseStr(passwordMac)
	return rsa.Encrypt(passwordEncode)
}
