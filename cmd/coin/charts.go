package main

// Code generated from files in charts subdirectory. DO NOT EDIT.
//
//go:generate go run charts/embed.go

var charts = map[string][]byte{}

func init() {
	for file, encoded := range map[string]string{
    	"totals.css": "H4sIAAAAAAAA/5TNQQrCMBCF4X1O8S6Qgi7T0wT7YgfGqWQGbBXvLurGldDt/8F7g9aNHcE18EgA0ES14DZLcPyEt+Vqp3npBbTpW9tikVu9iG4FXs2zs0v7QZc7Cw7H6zqmZ0qD8kyb/l7tHX0FAAD//4Ybv82+AAAA",
    	"totals.js": "H4sIAAAAAAAA/6RVS4/bNhC+61cMWCAg1zIjZRsgWMMp0vbQQ4r2kJuhA1eiZSE0ZZC0LWWx/73gQ7Zorbfrdg9rPmY+fvP4RgemoGKGwRKqe1rqw99MaY6re6q54KXBaPeTvUeEGt4ZTFJrx/am/dbvOEkSC6Da4x+8qTcGlpB/TBMAgC1TdSNhCU+m3T2cTWYfshSUXT2AXT62xrRbvxZ8bR7gY/bsIY5NZTawhE9ZBvMASK3NeeeAvPVmYGDpUsFlbTZ3Z2azwcW0u/PGv77wYehD7dMwhP7YVj0ilO12XFYY6UONiHuLMmMURo4gSgPRWURxFlGM3DxTlA6UX6MWMlxumLKx6UN9onNBxigm9bpVW5SC3whmOEYTXihFl08igkhIQhdSUDLBvzaSM4UJVUzWHK+yEGpBnGk/Mv2VySoy9LEFyx8jy79U1UgmXIuVG77lvzHD61b1eUYWnsKXrtHeg3WN/tbucBdejG++8rXBfUiRYD1X4VIbVn7HhH7nvcauH8pW7LdSUy2akuOcEHdMkqSnVbtljcTef5UVdMt2eL2XpWlaiSsCT6C42SsJFXVYvzPDF/BMyCLpBm8bcnVPt6wbgPxPaESYQ16k8DLqKi/gmRSESsttKITzh6UvfWjJL0JgRN3NUHzLKDwZTrg03BbtSp+UgmlteySC0aYXHKN1IwQa80yhGVH9gRtiI3c+iQMYM1O8NBGvl+M9IUyojgE82x5dyVqPR9U4kwpu3TW3DlerrJjYD0q+6pMXBObXvE+C7ukjk5UDw9Y+t8WcpMkO0v+VpjGAm8ojgCd3bP8OVgurvJhbzovT8RGWgP9kZkNFW+cZPhCYQU7gDj6djQKHA3yGDN69g86afYYj/ALO1ai9LO3ZAyB0U/Lzgsw/XHq8scqzKL939+9/dkCJl8i/9budGGD/zefdYFIyIbCbN7ZSN+P0EU4/4Dj58prLKozsSL3uIqr/yxPqRjVHsDd/EjISXB151yOn2UZsE+UZeR8T9ZNtng/TwBF4RcpxW6RN1BjHu2YiK9sV2dtlerxBlv9l6FkpxzG+pkN4Oul4oo6LLIyTMLNfwWkaJo1v2ST/BAAA//8moBBqvQkAAA==",
	} {
		charts[file] = decode(file, encoded)
	}
}
