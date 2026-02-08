module github.com/badele/ansi-compositor

go 1.25.4

require (
	github.com/alecthomas/kong v1.13.0
	github.com/badele/splitans v0.0.0-20260109231556-b5aed8e39501
	gopkg.in/yaml.v3 v3.0.1
)

require golang.org/x/text v0.31.0 // indirect

replace github.com/badele/splitans => /home/badele/ghq/github.com/badele/splitans
