package sfuzz

import (
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
)

var strList = []any{
	"", " ", "  ", "   ",
	".", "..", "...",
	`\`, `\\`,
	"undefined", "undef", "null", "NULL", "(null)", "nil", "NIL",
	"true", "false", "True", "False", "TRUE", "FALSE", "None",
}

var numList = []any{
	math.MaxInt64,
	math.MinInt64,
	fmt.Sprintf("%d00", uint64(math.MaxUint64)),
	fmt.Sprintf("%d00", math.MinInt64),
	"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999",
	0, -0, -0.00, 0.00, -1.00, -1.0, 1e23, -1e23,
	"inf", "Infinity", "-Infinity", "NaN", "INF",
	"0xffffffffffffffff", 0x0,
}

var uidList = []any{
	uuid.Nil.String(),
	fmt.Sprintf("%s%s", uuid.New().String(), uuid.New().String()),
	strings.ReplaceAll(uuid.New().String(), "-", "."),
	strings.ReplaceAll(uuid.New().String(), "-", ""),
}
