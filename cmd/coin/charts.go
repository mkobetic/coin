package main

// Code generated from files in charts subdirectory. DO NOT EDIT.
//
//go:generate go run charts/embed.go

var charts = map[string][]byte{}

func init() {
	for file, encoded := range map[string]string{
    	"totals.css": "H4sIAAAAAAAA/5TNQQrCMBCF4X1O8S6Qgi7T0wT7YgfGqWQGbBXvLurGldDt/8F7g9aNHcE18EgA0ES14DZLcPyEt+Vqp3npBbTpW9tikVu9iG4FXs2zs0v7QZc7Cw7H6zqmZ0qD8kyb/l7tHX0FAAD//4Ybv82+AAAA",
    	"totals.js": "H4sIAAAAAAAA/6RVQW/bNhS++1c8cEBBxjIrNStQxHCHbjvs0GE79GYIKCPRMlGaEkhalhrkvw8kJVu046xJc4gl8r2P33v8vqeve8PBWC0K+3U5m7VMQ8ksgxWUt7Qw7b9MG47LW2q45IXFqPnF7SNCLe8sJomLY3tbf+kbTgYEXR/+4qLaWlhB9j6ZAQDsmK6EghU82Lq5O4XM36UJaPd0B+7xvra23oVnyTf2Dt6njwHiIEq7hRV8SFNYDIDUxZzePFCI3o4MHF8quars9ubEbD6m2Lo5vYTThzJMW4U+jLXf12WPCGVNw1WJkWkrRPxZlFmrMfIEUTIQnUcU5xHFKC0wRclI+TlqY4uLLdOuONNWRz5nbKxmymxqvUMJhBfJLMfoghhK0PmZiKDxpG7oQcEk/ywUZxoTqpmqOF6nQ605WfrYfhL7O1NlFBmqG0O/T0L/0aVQTHqZFVu+438wy6ta91k6RHefOmFCBuuE+VI3uBvPjLc+843F/Uhesp7rYddYVnzDhH7jvcFeFEUt9ztlqJGi4DgjxC+73J6W9Y4JhQPAOs3pjjV4s1eFFbXCJYEH0NzutYKSerA/meVLeCRkOevGbFd2eUt3rBuBws8gR1hAlifwNOo6y+GR5IQqRy4qCFbh/gdhfpISI+p3RgU4RsORwwpXlrubuyKWQjJjnFAiGGN7yTHaCCnRlGcCYkL1OxbEVb70STOPMKWmeWEjYk8X7CCe5joFCHR7dKVtPZ5cBzlBhrTuWlqHy3WaX8SPhr6ak+UEFteyj77u6T1TpQfDLj5zt3nRJjdQf6pNUwA/nScAD37Z/bXODessXzjOy+PyAVaA/2Z2S2VdZSluCcwhI3ADH05BA4cWPkIKb95A58I+wgF+A59q9V4Vbu0OEHpR87OcLN6dZ/zgLc+j/t7cvv01qHEWTPJ/indTA9y/xaIbQwomJfZD51U4fYTTjzjewLziqhwmd+RfvxEJ4Okh9UI/R7Av/jKk/kPgfe3Ye5UcxxtxMspS8jZmGobbIqjcH/+Mk2NVJCLSxeFGXLjKiSL9cZceXuDK1w298xqfsyE8HG18YY6zLkybMHcfwss2XOieLGf/BQAA///xj2RY0QkAAA==",
	} {
		charts[file] = decode(file, encoded)
	}
}
