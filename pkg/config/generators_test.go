package config_test

import (
	"testing"

	"github.com/squarefactory/cloud-burster/pkg/config"
	"github.com/stretchr/testify/suite"
)

type GeneratorsTestSuite struct {
	suite.Suite
}

func (suite *GeneratorsTestSuite) TestSplitCommaOutsideOfBrackets() {
	tests := []struct {
		input    string
		expected []string
		title    string
	}{
		{
			input: "cn[1,2]cn[2,5-7],cn3,cn[4,5]",
			expected: []string{
				"cn[1,2]cn[2,5-7]",
				"cn3",
				"cn[4,5]",
			},
			title: "Positive test",
		},
		{
			input:    "",
			expected: []string{""},
			title:    "Empty test",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Act
			actual := config.SplitCommaOutsideOfBrackets(tt.input)

			// Assert
			suite.Equal(tt.expected, actual)
		})
	}
}

func (suite *GeneratorsTestSuite) TestExpandBrackets() {
	tests := []struct {
		input    string
		expected []string
		title    string
	}{
		{
			input: "cn[1,2]cn[2,5-7]",
			expected: []string{
				"cn1cn2",
				"cn1cn5",
				"cn1cn6",
				"cn1cn7",
				"cn2cn2",
				"cn2cn5",
				"cn2cn6",
				"cn2cn7",
			},
			title: "Positive test",
		},
		{
			input:    "",
			expected: []string{},
			title:    "Empty test",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Act
			actual := config.ExpandBrackets(tt.input)

			// Assert
			suite.Equal(tt.expected, actual)
		})
	}
}

func (suite *GeneratorsTestSuite) TestParseRangeList() {
	tests := []struct {
		input    string
		expected []int
		title    string
	}{
		{
			input:    "1,2,6-9",
			expected: []int{1, 2, 6, 7, 8, 9},
			title:    "Positive test",
		},
		{
			input:    "1,2,6-9a",
			expected: []int{1, 2},
			title:    "Bad range input",
		},
		{
			input:    "1,2,6-9-10",
			expected: []int{1, 2},
			title:    "Bad range input 2",
		},
		{
			input:    "1a,2,6-9",
			expected: []int{2, 6, 7, 8, 9},
			title:    "Bad digit input",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Act
			actual := config.ParseRangeList(tt.input)

			// Assert
			suite.Equal(tt.expected, actual)
		})
	}
}

func (suite *GeneratorsTestSuite) TestGenerateHostsFromGroupHost() {
	hostTemplate := config.Host{
		DiskSize:   50,
		FlavorName: "flavor",
		ImageName:  "image",
	}
	tests := []struct {
		input    config.GroupHost
		expected []config.Host
		isError  bool
		title    string
	}{
		{
			input: config.GroupHost{
				NamePattern:  "cn[1-5]",
				IPcidr:       "172.20.0.0/20",
				HostTemplate: hostTemplate,
			},
			isError: false,
			expected: []config.Host{
				{
					Name:       "cn1",
					DiskSize:   hostTemplate.DiskSize,
					FlavorName: hostTemplate.FlavorName,
					ImageName:  hostTemplate.ImageName,
					IP:         "172.20.0.1",
				},
				{
					Name:       "cn2",
					DiskSize:   hostTemplate.DiskSize,
					FlavorName: hostTemplate.FlavorName,
					ImageName:  hostTemplate.ImageName,
					IP:         "172.20.0.2",
				},
				{
					Name:       "cn3",
					DiskSize:   hostTemplate.DiskSize,
					FlavorName: hostTemplate.FlavorName,
					ImageName:  hostTemplate.ImageName,
					IP:         "172.20.0.3",
				},
				{
					Name:       "cn4",
					DiskSize:   hostTemplate.DiskSize,
					FlavorName: hostTemplate.FlavorName,
					ImageName:  hostTemplate.ImageName,
					IP:         "172.20.0.4",
				},
				{
					Name:       "cn5",
					DiskSize:   hostTemplate.DiskSize,
					FlavorName: hostTemplate.FlavorName,
					ImageName:  hostTemplate.ImageName,
					IP:         "172.20.0.5",
				},
			},
			title: "Positive test",
		},
		{
			input: config.GroupHost{
				NamePattern:  "cn[1-2000]",
				IPcidr:       "172.20.0.0/24",
				HostTemplate: hostTemplate,
			},
			isError:  true,
			expected: []config.Host{},
			title:    "Not enough IP",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.title, func() {
			// Act
			actual, err := config.GenerateHostsFromGroupHost(tt.input)

			// Assert
			if tt.isError {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
			suite.Equal(tt.expected, actual)
		})
	}
}

func TestGeneratorsTestSuite(t *testing.T) {
	suite.Run(t, &GeneratorsTestSuite{})
}
