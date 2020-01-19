package main

// Code generated from files in charts subdirectory. DO NOT EDIT.
//
//go:generate go run charts/embed.go

var charts = map[string][]byte{}

func init() {
	for file, encoded := range map[string]string{
    	"totals.js": "H4sIAAAAAAAA/5xVwW7bOBC96ysGXAQg1zIjr7HAwoaxSNtjmvaQS2H4wEi0TYSiBHGsSgn87wUpypYcu4fyYIvke2/eaGagWlSQCRSwgmzOU1t/F5WVNJtzK7VMkZLyL3dPGEfZIGWxw4kDFs9tKVkUOYFcVDtlYAXvWJQLmCcxVGq3xwX8k8TwUiAWefes5RYX8G9yjCMAgJ8qwz2s4L8kgWmQ4Q5z3nmhDr2X7tk5FSi4lmaH+7/nCUx6LBbledOFXXYObb3rMuyzeimyljAuylKajBJb7wjzQbhArCjxzkgcHE5G3iYjbyNaZ5HEvdffWeuJwcKFAayEsduiykkM3UYLlJR88EJichmGMMJC4k1IOxVaPiojRUUZr4TZSbpOQnob5qHtAPpJmGwE7PIJyLcB8luVKSO075h0L3P5WaDcFVU7S9iys/DQKNsxRKPsc1HSJkQc3zzKLdI29JQWrazCpUWRvlLGX2VrqS9+WuhDbiy3WqWSzhjzxyyKWp4VuVCGdvx1suG5KOn2YFJUhaEZg3eoJB4qAxn3Wl8EyiUcGVtGTc92KWdznoumF+r+QtfBFGabGK6rrmcbOLIN48Z56wvh+bBynRia8EFrSrg/70vv/ISA4UQalK5kN7ok1cJa1yEjGYutlpRsldZk6DIGNTD6RhVzeS+jyLOHtiqZ4sjU9VQd/brPgcDJbEtuvLKWDkrBTqInYnOL2NBsnWyuMPrxvcmabRhMb/NPc9zyF2EyL0cdY+ZeF/gqXpbkY1FcU4P7mU6bMygVWlM/FH+o1V5otb1WdH8/zPjd7d069ftoFOpYnSFuqS1QBasVJAxGF265HsZuHFHlsvtMkLsf93dfCaM1u8SHmOj0zEFr+B9qWAAOcUeQ2srrsdznZFLfUFX2STxRw4KoGYn2m6P3dGS/AgAA//9/fonc4wYAAA==",
	} {
		charts[file] = decode(file, encoded)
	}
}
