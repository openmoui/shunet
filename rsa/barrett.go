package rsa

import "shunet/bigint"

type barrettMu struct {
	modulus bigint.BigInt
	k       int
	mu      bigint.BigInt
	bkplus1 bigint.BigInt
}

func newBarrettMu(m bigint.BigInt) (t barrettMu) {
	t.modulus = bigint.BiCopy(m)
	t.k = t.modulus.BiHighIndex() + 1
	b2k := bigint.NewBigInt()
	b2k.Digits[2*t.k] = 1
	t.mu, _ = bigint.BiDivideModulo(b2k, t.modulus)
	t.bkplus1 = bigint.NewBigInt()
	t.bkplus1.Digits[t.k+1] = 1
	return
}

// BarrettMu_modulo 计算模运算
func (brtm *barrettMu) BarrettMu_modulo(x bigint.BigInt) bigint.BigInt {
	q1 := x.BiDivideByRadixPower(brtm.k - 1)
	q2 := bigint.BiMultiply(q1, brtm.mu)
	q3 := q2.BiDivideByRadixPower(brtm.k + 1)
	r1 := x.BiModuloByRadixPower(brtm.k + 1)
	r2term := bigint.BiMultiply(q3, brtm.modulus)
	r2 := r2term.BiModuloByRadixPower(brtm.k + 1)
	r := bigint.BiSubtract(r1, r2)
	if r.IsNeg {
		r = bigint.BiAdd(r, brtm.bkplus1)
	}
	for bigint.BiCompare(r, brtm.modulus) >= 0 {
		r = bigint.BiSubtract(r, brtm.modulus)
	}
	return r
}

// MultiplyMod 计算 (x * y) % modulus
func (brtm *barrettMu) MultiplyMod(x, y bigint.BigInt) bigint.BigInt {
	// 乘法运算
	xy := bigint.BiMultiply(x, y)
	// 进行模运算
	return brtm.BarrettMu_modulo(xy)
}

// BarrettMu_powMod 计算 (x^y) % modulus
func (brtm *barrettMu) BarrettMu_powMod(x, y bigint.BigInt) bigint.BigInt {
	result := bigint.NewBigInt()
	result.Digits[0] = 1
	a := bigint.BiCopy(x)
	k := bigint.BiCopy(y)
	for {
		if (k.Digits[0] & 1) != 0 {
			result = brtm.MultiplyMod(result, a)
		}
		k = k.BiShiftRight(1)
		if k.Digits[0] == 0 && k.BiHighIndex() == 0 {
			break
		}
		a = brtm.MultiplyMod(a, a)
	}
	return result
}
