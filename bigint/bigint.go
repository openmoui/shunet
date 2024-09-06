package bigint

import (
	"math"
)

type BigInt struct {
	Digits []int
	IsNeg  bool
}

var (
	biRadixBits    = 16 //BigInt中进制的位数，默认为2^16进制
	biHalfRadix    = int(uint16(biRadix) >> 1)
	biRadix        = 1 << biRadixBits // BigInt的进制
	biRadixSquared = biRadix * biRadix
	maxDigitVal    = biRadix - 1 // 一位的最大数字，这里的一位不是指比特位，而是指Digits的一个元素的最大数字
	bitsPerDigit   = biRadixBits

	BigZero BigInt = NewBigInt()
	BigOne  BigInt = func() BigInt {
		tmp := NewBigInt()
		tmp.Digits[0] = 1
		return tmp
	}()
)

func NewBigInt() BigInt {
	return BigInt{Digits: make([]int, 130)}
}

func (bi *BigInt) BiHighIndex() int {
	var result = len(bi.Digits) - 1
	for result > 0 && bi.Digits[result] == 0 {
		result--
	}
	return result
}

func (bi *BigInt) BiNumBits() int {
	n := bi.BiHighIndex() // 找到最高有效字节的索引
	if n < 0 {
		return 0
	}
	d := bi.Digits[n] // 获取最高有效字节
	m := (n + 1) * bitsPerDigit
	result := m

	for result > m-8 { // 在当前字节的位中查找最高位的有效位数
		if (d & 0x80) != 0 { // 检查最高位是否为1
			break
		}
		d <<= 1
		result--
	}
	return result
}

func BiAdd(x, y BigInt) BigInt {
	result := NewBigInt()
	carry := 0
	for i := 0; i < len(x.Digits); i++ {
		sum := x.Digits[i] + y.Digits[i] + carry
		if sum >= biRadix {
			carry = 1
			sum -= biRadix
		} else {
			carry = 0
		}
		result.Digits[i] = sum
	}
	return result
}

func BiSubtract(x, y BigInt) BigInt {
	var result BigInt
	if x.IsNeg != y.IsNeg {
		// Signs are different, so we perform an addition
		yCopy := BiCopy(y)
		yCopy.IsNeg = !yCopy.IsNeg
		result = BiAdd(x, yCopy)
	} else {
		// Signs are the same, so we perform a subtraction
		result = NewBigInt()
		carry := 0
		for i := 0; i < len(x.Digits); i++ {
			diff := x.Digits[i] - y.Digits[i] + carry
			if diff < 0 {
				diff += biRadix
				carry = -1
			} else {
				carry = 0
			}
			result.Digits[i] = diff
		}

		// Adjust the sign if necessary
		if carry == -1 {
			carry = 0
			for i := 0; i < len(x.Digits); i++ {
				diff := 0 - result.Digits[i] + carry
				if diff < 0 {
					diff += biRadix
					carry = -1
				} else {
					carry = 0
				}
				result.Digits[i] = diff
			}
			result.IsNeg = !x.IsNeg
		} else {
			result.IsNeg = x.IsNeg
		}
	}
	return result
}

// BiMultiply 执行两个大整数的乘法运算
func BiMultiply(x, y BigInt) BigInt {
	result := NewBigInt()
	n := x.BiHighIndex()
	t := y.BiHighIndex()

	for i := 0; i <= t; i++ {
		c := 0
		k := i
		for j := 0; j <= n; j, k = j+1, k+1 {
			uv := result.Digits[k] + x.Digits[j]*y.Digits[i] + c
			result.Digits[k] = uv & maxDigitVal
			c = uv >> biRadixBits
		}
		result.Digits[i+n+1] = c
	}

	// 设置结果的符号位
	result.IsNeg = x.IsNeg != y.IsNeg
	return result
}

func (bi *BigInt) BiMultiplyDigit(y int) BigInt {
	result := NewBigInt()
	carry := 0
	for i := 0; i < len(bi.Digits); i++ {
		product := bi.Digits[i]*y + carry
		result.Digits[i] = product & maxDigitVal
		carry = product >> bitsPerDigit
	}
	return result
}

func (bi *BigInt) BiMultiplyByRadixPower(n int) BigInt {
	result := NewBigInt()
	for i := 0; i < len(bi.Digits)-n; i++ {
		result.Digits[i+n] = bi.Digits[i]
	}
	return result
}

func BiDivideModulo(x, y BigInt) (BigInt, BigInt) {
	nb := x.BiNumBits()
	tb := y.BiNumBits()
	origYIsNeg := y.IsNeg
	var q = NewBigInt()
	var r = NewBigInt()

	if nb < tb {
		if x.IsNeg {
			q = BiCopy(BigOne)
			q.IsNeg = !y.IsNeg
			x.IsNeg = false
			y.IsNeg = false
			r = BiSubtract(y, x)
			x.IsNeg = true
			y.IsNeg = origYIsNeg
		} else {
			q = NewBigInt()
			r = BiCopy(x)
		}
		return q, r
	}

	q = NewBigInt()
	r = BiCopy(x)

	t := int(math.Ceil(float64(tb)/float64(bitsPerDigit))) - 1
	lambda := 0
	for y.Digits[t] < biHalfRadix {
		y = y.BiShiftLeft(1)
		lambda++
		tb++
		t = int(math.Ceil(float64(tb)/float64(bitsPerDigit))) - 1
	}

	r = r.BiShiftLeft(lambda)
	nb += lambda
	n := int(math.Ceil(float64(nb)/float64(bitsPerDigit))) - 1

	b := y.BiMultiplyByRadixPower(n - t)
	for BiCompare(r, b) != -1 {
		q.Digits[n-t]++
		r = BiSubtract(r, b)
	}

	for i := n; i > t; i-- {
		ri := 0
		if i < len(r.Digits) {
			ri = r.Digits[i]
		}
		ri1 := 0
		if i-1 < len(r.Digits) {
			ri1 = r.Digits[i-1]
		}
		ri2 := 0
		if i-2 < len(r.Digits) {
			ri2 = r.Digits[i-2]
		}
		yt := 0
		if t < len(y.Digits) {
			yt = y.Digits[t]
		}
		yt1 := 0
		if t-1 < len(y.Digits) {
			yt1 = y.Digits[t-1]
		}
		if ri == yt {
			q.Digits[i-t-1] = maxDigitVal
		} else {
			q.Digits[i-t-1] = (ri*biRadix + ri1) / yt
		}

		c1 := q.Digits[i-t-1] * (yt*biRadix + yt1)
		c2 := ri*biRadixSquared + ri1*biRadix + ri2
		for c1 > c2 {
			q.Digits[i-t-1]--
			c1 = q.Digits[i-t-1] * (yt*biRadix + yt1)
			c2 = ri*biRadixSquared + ri1*biRadix + ri2
		}

		b = y.BiMultiplyByRadixPower(i - t - 1)
		r = BiSubtract(r, b.BiMultiplyDigit(q.Digits[i-t-1]))
		if r.IsNeg {
			r = BiAdd(r, b)
			q.Digits[i-t-1]--
		}
	}
	r = r.BiShiftRight(lambda)
	q.IsNeg = x.IsNeg != origYIsNeg
	if x.IsNeg {
		if origYIsNeg {
			q = BiAdd(q, BigOne)
		} else {
			q = BiSubtract(q, BigOne)
		}
		y = y.BiShiftRight(lambda)
		r = BiSubtract(y, r)
	}
	if r.Digits[0] == 0 && r.BiHighIndex() == 0 {
		r.IsNeg = false
	}

	return q, r
}

func (bi *BigInt) BiDivideByRadixPower(n int) BigInt {
	result := NewBigInt()
	for i := n; i < len(bi.Digits); i++ {
		result.Digits[i-n] = bi.Digits[i]
	}
	return result
}

func (bi *BigInt) BiModuloByRadixPower(n int) BigInt {
	result := NewBigInt()
	for i := 0; i < n && i < len(bi.Digits); i++ {
		result.Digits[i] = bi.Digits[i]
	}
	return result
}

func (bi *BigInt) BiShiftLeft(n int) BigInt {
	result := NewBigInt()
	digitCount := n / bitsPerDigit
	for i := 0; i < len(bi.Digits)-digitCount; i++ {
		result.Digits[i+digitCount] = bi.Digits[i]
	}
	bits := n % bitsPerDigit
	rightBits := bitsPerDigit - bits
	for i := len(result.Digits) - 1; i > 0; i-- {
		result.Digits[i] = ((result.Digits[i] << bits) & maxDigitVal) | ((result.Digits[i-1] & ((1 << bits) - 1<<rightBits)) >> rightBits)
	}
	result.Digits[0] = (bi.Digits[0] << bits) & maxDigitVal
	result.IsNeg = bi.IsNeg
	return result
}

func (bi *BigInt) BiShiftRight(n int) BigInt {
	result := NewBigInt()
	digitCount := n / bitsPerDigit
	for i := 0; i < len(bi.Digits)-digitCount; i++ {
		result.Digits[i] = bi.Digits[i+digitCount]
	}
	bits := n % bitsPerDigit
	leftBits := bitsPerDigit - bits
	for i := 0; i < len(result.Digits)-1; i++ {
		result.Digits[i] = (result.Digits[i] >> bits) | ((result.Digits[i+1] & ((1 << bits) - 1)) << leftBits)
	}
	result.Digits[len(result.Digits)-1] >>= bits
	result.IsNeg = bi.IsNeg
	return result
}

func BiCopy(src BigInt) BigInt {
	res := NewBigInt()
	copy(res.Digits, src.Digits)
	res.IsNeg = src.IsNeg
	return res
}

// BiCompare 比较两个 BigInt 的大小
func BiCompare(x, y BigInt) int {
	if x.IsNeg != y.IsNeg {
		if x.IsNeg {
			return -1
		}
		return 1
	}
	for i := len(x.Digits) - 1; i >= 0; i-- {
		if x.Digits[i] < y.Digits[i] {
			if x.IsNeg {
				return 1
			}
			return -1
		} else if x.Digits[i] > y.Digits[i] {
			if x.IsNeg {
				return -1
			}
			return 1
		}
	}
	return 0
}
