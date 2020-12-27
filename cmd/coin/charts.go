package main

// Code generated from files in charts subdirectory. DO NOT EDIT.
//
//go:generate go run charts/embed.go

var charts = map[string][]byte{}

func init() {
	for file, encoded := range map[string]string{
    	"totals.css": "H4sIAAAAAAAA/5TNQQrCMBCF4X1O8S6Qgi7T0wT7YgfGqWQGbBXvLurGldDt/8F7g9aNHcE18EgA0ES14DZLcPyEt+Vqp3npBbTpW9tikVu9iG4FXs2zs0v7QZc7Cw7H6zqmZ0qD8kyb/l7tHX0FAAD//4Ybv82+AAAA",
    	"totals.js": "H4sIAAAAAAAA/6RVQW/bNhS++1c8cEBBxhIrJStQxHCHbjvs0GE79GYIKCPRMlGaMkhalhrkvw8kJVu046xpc4gl8r2P33v8vqcve8PBWC1K+2Uxm7VMQ8UsgyVUd7Q07b9MG46rO2q45KXFaPeL20eEWt5ZTBIXx/a2+dzvOBkQdHP4i4t6Y2EJ+btkBgCwZboWCpbwaJvd/SlkfpsloN3TPbjHh8baZhueJV/be3iXPQWIg6jsBpbwPssgHQCpizm9eaAQvRkZOL5UclXbzc2J2XxMsc3u9BJOH8owbR36MNb+0FQ9IpTtdlxVGJm2RsSfRZm1GiNPECUD0XlEcR5RjNICU5SMlF+iNra43DDtijNtfeRzxsZqpsy60VuUQHiRzHKMLoihBJ2fiQgaT+qGHpRM8k9CcaYxoZqpmuNVNtRakIWP7SexvzNVRZGhujH02yT0H10JxaSXWbnhW/4Hs7xudJ9nQ3T3sRMmZLBOmM/NDnfjmfHWJ762uB/JS9ZzPeway8qvmNCvvDfYi6Js5H6rDDVSlBznSZoT4ndcek+rZsuEwgFjlRV0y3Z4vVelFY3CFYFH0NzutYKKerw/meULeCJkMevGbFd5dUe3rBuBws+gSEghLxJ4HnWVF/BECkKV4xfVBMsggUGbH6XEiPqdUQSO0XDksMKV5e7yruillMwYp5UIxthecozWQko05ZmAmFD9hgVxlS980swjTKlpXtqI2PMFO4jnuU4BAt0eXWlbjyfXQU6QIa27ltbhapUVF/Gjp5/NyRKflhcE0msAR3f39IGpyuNhF5+7C73olBurP9WpKYCf0ROAR7/s/lrniVVepI7z4rh8gCXgv5ndUNnUeYZbAnPICdzA+1PQwKGFD5DBmzfQubAPcIDfwKdavVelW7sHhF7V/7wg6e15xnde9Dzq783d21+DIGfBJ/8nejc7wP1L024MKZmU2I+eH8LpI5x+xPEe5jVX1TC/Iwv7jUgAV0fVK10dIb/6E5H5L4J3tyvAC+U45IhTUp6RtzHZMOLSWy90f/wLfo6FkYhIGocbcWEsp4vs+716eIUxf2z0ndf4khPh8ejkC3+cdWHahLn7Il624UL6ZDH7LwAA///1/5Hn2gkAAA==",
	} {
		charts[file] = decode(file, encoded)
	}
}
