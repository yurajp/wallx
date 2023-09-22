package ascod

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type Bigint struct {
	Val big.Int
}

type PubKey struct {
	E Bigint
	N Bigint
}

func (pk PubKey) String() string {
  return fmt.Sprintf("Pub %s %s", pk.E.String(), pk.N.String())
}

type PrivKey struct {
	D Bigint
	N Bigint
}

func (pk *PrivKey) String() string {
  return fmt.Sprintf("Priv %s %s", pk.D.Val.String(), pk.N.Val.String())
}

func (b Bigint) MarshalJSON() ([]byte, error) {
	return []byte(b.Val.String()), nil
}

func (b *Bigint) UnmarshalJSON(p []byte) error {
	var v big.Int
	err := v.UnmarshalText(p)
	if err != nil {
		return fmt.Errorf("%s not valid", string(p))
	}
	b.Val = v
	return nil
}

func (b Bigint) String() string {
	return b.Val.String()
}

type KeyResp struct {
	Rand string
	Pub  PubKey
}

func NewKeyResp(rn string, pb PubKey, pr *PrivKey) *KeyResp {
  ern := SrvEncodeString(rn, pr)
  return &KeyResp{ern, pb}
}

func (kr *KeyResp) GetPubAndCheck(rn string) (PubKey, bool) {
  drn := ClDecodeString(kr.Rand, kr.Pub)
  return kr.Pub, drn == rn
}

func pRoot(f big.Int) (big.Int, big.Int, error) {
	che := []int64{3, 5, 7, 11, 13, 17, 19, 29}
	var ch int64 = 31
	fi := f.Int64()
	var mli int64 = 2
  Multi:
	for mli < 9 {
		for _, c := range che {
			if (mli*fi+1)%c == 0 {
				ch = c
				break Multi
			}
		}
		mli += 1
	}
	if ch > 29 {
		return *big.NewInt(0), *big.NewInt(0),
			errors.New("unsuccess when choose root")
	}
	return *big.NewInt(mli), *big.NewInt(ch), nil
}

func GenerateKeys() (PubKey, *PrivKey, error) {
	p, _ := rand.Prime(rand.Reader, 16)
	q, _ := rand.Prime(rand.Reader, 16)
	one := big.NewInt(1)
	var fi, n, p1, q1 big.Int
	p1.Sub(p, one)
	q1.Sub(q, one)
	n.Mul(p, q)
	fi.Mul(&p1, &q1)
	var d, dw, de big.Int
	mlt, e, err := pRoot(fi)
	if err != nil {
		return PubKey{}, &PrivKey{}, fmt.Errorf("ERROR ocured: %s", err)
	}
	dw.Mul(&mlt, &fi)
	de.Add(&dw, one)
	d.Div(&de, &e)
	pbe := PubKey{}
	pbe.E.Val = e
	pbe.N.Val = n
	prv := PrivKey{}
	prv.D.Val = d
	prv.N.Val = n
	return pbe, &prv, nil
}

func srvDecField(crp string, priv *PrivKey) string {
	var bnum, dec big.Int
	bnum.UnmarshalText([]byte(crp))
	dec.Exp(&bnum, &priv.D.Val, &priv.N.Val)
	return dec.String()
}

func clDecField(cryp string, pub PubKey) string {
	var bnum, dec big.Int
	bnum.UnmarshalText([]byte(cryp))
	dec.Exp(&bnum, &pub.E.Val, &pub.N.Val)
	return dec.String()
}

func clEncRune(r rune, pb PubKey) string {
  var er big.Int
  er.UnmarshalText([]byte(fmt.Sprintf("%v", r)))
	var cr big.Int
	cr.Exp(&er, &pb.E.Val, &pb.N.Val)
	return cr.String()
}

func srvEncRune(r rune, pr *PrivKey) string {
  var er big.Int
  er.UnmarshalText([]byte(fmt.Sprintf("%v", r)))
	var cr big.Int
	cr.Exp(&er, &pr.D.Val, &pr.N.Val)
	return cr.String()
}

func GeneratePassword(n int) string {
	gpw := make([]rune, n)
	for i := 0; i < n; i++ {
		r, _ := rand.Int(rand.Reader, big.NewInt(75))
		if r.Int64() == 44 {
			continue
		}
		gpw[i] = rune(r.Int64() + int64(48))
	}
	return string(gpw)
}

func HashStr(ps string) string {
	h := sha256.New()
	h.Write([]byte(ps))
	hh := h.Sum(nil)
	hhs := base64.StdEncoding.EncodeToString(hh)
	return string(hhs)
}

func SrvEncodeString(s string, pr *PrivKey) string {
  bd := strings.Builder{}
  for _, r := range s {
    er := srvEncRune(r, pr)
    bd.WriteString(er)
    bd.WriteString(" ")
  }
  return bd.String()
}

func ClDecodeString(s string, pb PubKey) string {
  bd := strings.Builder{}
  for _, er := range strings.Fields(s) {
    dr := clDecField(er, pb)
    di, _ := strconv.Atoi(dr)
    bd.WriteRune(rune(di))
  }
  return bd.String()
}

func ClEncodeString(s string, pb PubKey) string {
  bd := strings.Builder{}
  for _, r := range s {
    er := clEncRune(r, pb)
    bd.WriteString(er)
    bd.WriteString(" ")
  }
  return bd.String()
}

func SrvDecodeString(es string, pr *PrivKey) string {
  bd := strings.Builder{}
  for _, s := range strings.Fields(es) {
    dr := srvDecField(s, pr)
    di, _ := strconv.Atoi(dr)
    bd.WriteRune(rune(di))
  }
  return bd.String()
}

func GenRandom() (string, error) {
	b, err := rand.Int(rand.Reader, big.NewInt(999999))
	if err != nil {
		return "", fmt.Errorf("Cannot generate random: %w", err)
	}
	return b.String(), nil
}

func ClEncConfirm(pb PubKey) string {
  return ClEncodeString("OK", pb)
}

func IsClConfirmed(ok string, pb PubKey) bool {
  dcs := ClDecodeString(ok, pb) 
  return dcs == "OK"
}

