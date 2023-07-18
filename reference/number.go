package reference

import "strconv"

type Number string

func (n Number) Next() Number {
	base := n[0 : len(n)-1]
	baseN, err := strconv.ParseInt(string(base), 10, 64)
	if err != nil {
		panic(err)
	}
	next := strconv.FormatInt(baseN+1, 10)

	weights := []int{7, 3, 1}
	sum := 0
	for i := 0; i < len(next); i++ {
		digit := int(next[len(next)-i-1] - '0')
		weight := weights[i%3]
		sum += digit * weight
	}
	checkDigit := (10 - (sum % 10)) % 10
	return Number(next + strconv.Itoa(checkDigit))
}
