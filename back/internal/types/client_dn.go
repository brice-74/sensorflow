package types

import "fmt"

type ClientDN struct {
	CN string
	OU string
	O  string
	L  string
	ST string
	C  string
}

func ParseClientDN(dn string) (*ClientDN, error) {
	var res = new(ClientDN)
	start := 0
	l := len(dn)

	for i := 0; i <= l; i++ {
		if i == l || dn[i] == ',' {
			if i <= start+2 {
				return nil, fmt.Errorf("invalid DN segment: %q", dn[start:i])
			}

			part := dn[start:i]
			start = i + 1

			var key, val string
			for j := 0; j < len(part); j++ {
				if part[j] == '=' {
					if j == 0 || j == len(part)-1 {
						return nil, fmt.Errorf("invalid key=value pair: %q", part)
					}
					key = part[:j]
					val = part[j+1:]
					break
				}
			}

			if key == "" || val == "" {
				return nil, fmt.Errorf("invalid key=value pair: %q", part)
			}

			for k := 0; k < len(val); k++ {
				if val[k] < 32 || val[k] == ',' || val[k] == '=' {
					return nil, fmt.Errorf("invalid DN value for %s: %q", key, val)
				}
			}

			switch key {
			case "CN":
				res.CN = val
			case "OU":
				res.OU = val
			case "O":
				res.O = val
			case "L":
				res.L = val
			case "ST":
				res.ST = val
			case "C":
				res.C = val
			default:
				return nil, fmt.Errorf("invalid DN key: %q", key)
			}
		}
	}

	return res, nil
}
