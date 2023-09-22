package symcod

import (
   "strconv"
   "math/rand"
)


func toBinar(r rune) []int {
	bn := strconv.FormatInt(int64(r), 2)
	res := []int{}
	for _, b := range bn {
		i, _ := strconv.Atoi(string(b))
		res = append(res, i)
	}
	return res
}

func xorR(a, b []int) rune {
	bs := ""
	for idx, e := range a {
		n := 1 - e ^ b[idx%len(b)]
		bs += strconv.Itoa(n)
	}
	res, _ := strconv.ParseInt(bs, 2, 64)
	return rune(res)
}

func keySeed(key string) int64 {
	keyStr := ""
	for _, rn := range key {
		keyStr += strconv.Itoa(int(rn))
	}
	if len(keyStr) > 16 {
		nn := len(keyStr) - 16
		keyStr = keyStr[nn:]
	}
	if len(keyStr) < 7 {
		keyStr += keyStr
	}
	keyInt, _ := strconv.Atoi(keyStr)
	return int64(keyInt)
}

func SymEncode(text string, key string) string {
	kar := []int{}
	for _, r := range key {
		kar = append(kar, toBinar(r)...)
	}
	enc := ""
	rand.Seed(keySeed(key))
	kw := len(kar)
	one := []int{1}
	for i, r := range []rune(text) {
		bit := toBinar(r)
		kc := append(kar[i%kw:], kar[:i%kw]...)
		kco := append(one, kc...)
		e := xorR(bit, kco)
		rn := rand.Intn(36)
		osc := 56 + rn*(1-(i%2)*2)
		enc += string(e + rune(osc))
	}
	return enc
}

func SymDecode(code string, key string) string {
	kar := []int{}
	for _, r := range key {
		kar = append(kar, toBinar(r)...)
	}
	dec := ""
	rand.Seed(keySeed(key))
	crar := []rune(code)
	kw := len(kar)
	one := []int{1}
	for i, ec := range crar {
		rn := rand.Intn(36)
		osc := 56 + rn*(1-(i%2)*2)
		rut := ec - rune(osc)
		bit := toBinar(rut)
		kc := append(kar[i%kw:], kar[:i%kw]...)
		kco := append(one, kc...)
		dg := xorR(bit, kco)
		dec += string(dg)
	}
	return dec
}
