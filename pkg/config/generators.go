package config

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/squarefactory/cloud-burster/logger"
	"github.com/squarefactory/cloud-burster/utils/cidr"
	"go.uber.org/zap"
)

var matchCommaOutsideOfBrackets = regexp.MustCompile(`(?:\[+[^\[\]]*\]+|[^,])+`)

// SplitCommaOutsideOfBrackets generates strings
func SplitCommaOutsideOfBrackets(pattern string) []string {
	res := matchCommaOutsideOfBrackets.FindAllString(pattern, -1)
	if res == nil {
		return []string{""}
	}
	return res
}

// ExpandBrackets generates strings based on brackets ranges or digits.
//
// cn[1,2-4] generates cn1, cn2, cn4 and cn3.
func ExpandBrackets(pattern string) []string {
	var out []string
	if pattern == "" {
		return []string{}
	}

	beginIdx := -1
	for idx, rune := range pattern {
		// Search for '['
		if rune == '[' {
			beginIdx = idx
		}

		// Search for ']'
		if rune == ']' && beginIdx != -1 && beginIdx <= idx {
			digits := ParseRangeList(pattern[beginIdx+1 : idx])

			// Add the generated name
			for _, digit := range digits {
				out = append(out, fmt.Sprintf("%s%d%s", pattern[:beginIdx], digit, pattern[idx+1:]))
			}

			break
		}
	}

	var merge []string
	for _, pattern := range out {
		names := ExpandBrackets(pattern)
		if len(names) == 0 {
			// Pattern is at the smallest factor
			merge = append(merge, pattern)
		} else {
			merge = append(merge, ExpandBrackets(pattern)...)
		}
	}

	return merge
}

// ParseRangeList converts a string containing comma separated digits and ranges into an array of digits.
//
// For example, "1,2-4" is [1,2,3,4].
func ParseRangeList(ranges string) []int {
	var digits []int
	digitsOrRanges := strings.Split(ranges, ",")

	for _, digitOrRange := range digitsOrRanges {
		// Check for digit
		digit, err := strconv.Atoi(digitOrRange)
		if err == nil {
			digits = append(digits, digit)
			continue
		}

		// Check for range
		r := strings.Split(digitOrRange, "-")
		if len(r) != 2 {
			// Is not a range
			continue
		}
		begin, errBegin := strconv.Atoi(r[0])
		end, errEnd := strconv.Atoi(r[1])
		if errBegin == nil && errEnd == nil {
			for digit := begin; digit <= end; digit++ {
				digits = append(digits, digit)
			}
			continue
		}
	}
	return digits
}

func GenerateHostsFromGroupHost(group GroupHost) ([]Host, error) {
	var out []Host

	// Generates names based on Name Pattern
	names := ExpandBrackets(group.NamePattern)

	// Generates IPs
	ipAddresses := cidr.Hosts(group.IPcidr)

	if len(names) > len(ipAddresses) {
		logger.I.Error(
			"not enough IP addresses in CIDR",
			zap.String("namePattern", group.NamePattern),
			zap.Int("len(namePattern)", len(names)),
			zap.String("ipCIDR", group.IPcidr),
			zap.Int("len(ipAddresses)", len(ipAddresses)),
		)
		return []Host{}, errors.New("not enough IP addresses in CIDR")
	}

	// Map the names into host
	for idx, name := range names {
		host := Host{
			Name:       name,
			DiskSize:   group.HostTemplate.DiskSize,
			FlavorName: group.HostTemplate.FlavorName,
			ImageName:  group.HostTemplate.ImageName,
			IP:         ipAddresses[idx],
		}
		out = append(out, host)
	}

	return out, nil
}
