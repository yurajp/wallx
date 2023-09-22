package ascod

import (
  "fmt"
  "time"
)

type Perfon struct {
  Times []time.Time
}

func NewPerfon() *Perfon {
  return &Perfon{}
}

func (pf *Perfon) Do() {
  pf.Times = append(pf.Times, time.Now())
  lnt := len(pf.Times)
  if lnt > 1 {
    dur := pf.Times[lnt-1].Sub(pf.Times[lnt-2])
    fmt.Printf("  ✈  %d: %d mks\n", lnt-1, dur.Microseconds())
  } else {
    fmt.Println("  ✈  ✷")
  }
  pf.Times[lnt -1] = time.Now()
}

func (pf *Perfon) Re() {
  pf.Times = append(pf.Times, time.Now())
}

func (pf *Perfon) Stop() {
  pf.Times = []time.Time{}
  fmt.Println("  ✈  ✗")
}
