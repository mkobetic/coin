package main

// Code generated from files in charts subdirectory. DO NOT EDIT.
//
//go:generate go run charts/embed.go

var charts = map[string][]byte{}

func init() {
	for file, encoded := range map[string]string{
    	"totals.css": "H4sIAAAAAAAA/9LLSaxMLVIoSa0oUajmUlBQUEjLzMmxUijPyCxJtQYLgOR0E/OSM/KLrBRS81KsuWq5uPRyUtNT81JwaqzlAgQAAP//796EBFoAAAA=",
    	"totals.js": "H4sIAAAAAAAA/6RVTW/jNhC961cM2AsZy1wp7gJFDLfYtuhpgfawN0MHRqJlYmnKIBlH2iD/veCHHDGK03X3kkjizOObN/PGJ6ahYZbBBpoVrc3pH6YNx82KGi55bTE6/uTOEaGW9xaT3MWxB9t9GY6cZJkDODDdCgUbeLLd8Q4+Fjlo0e7tHdwWOdx31naH8Cz5zrqA5zwDAHgUjd3DBn4pClhGGOpiXt48UIjec/fsmDLLqOSqtfubVQGLMdZ2x5eXcO06MDSnNlQ4VnXfNQMilB2PXDUYmVOLiL+EMms1Rp4ZyiPDRcJtkXBL0gJFlI9c36MWxav3TLuizKk903lFxmqmzK7TB5RDeJHMcoxmvFCOXl+JCCJRhD5KUDPJPwvFmcaEaqZajrdFLLUiPnSYhP7OVJMEhtpi5LdJ5N+6EYpJPz31nh/4H8zyttNDWZB1oPCpFyZksF6YL90R9/HG9OQz31k8RIkkG7iOh8ay+ism9CsfDPaDUHfy4aAMNVLUHJeE+M8kywbadAcmFA7526KiB3bEuwdVW9Ep3BB4As3tg1bQUI/1J7N8Dc+ErLN+zHYlNyt6YP0IFP7FCYQllFUOb6NuywqeSUWoctzGRvh82ITWx5H8JCVG1J+MzXeM4pXxC1eWu6ZdmJNaMmPcjCQwxg6SY7QTUqIpzxzEhOo3LIir3OdkHmDKTPPaJrzerveMMKM6BQhsB3RBtQFPuvFCKqb1l9J63GyLahY/OvliTlkRWF7KPht6oPdMNR4Mu/jSNXMmk9uRPyTTFMAv3AnAE5ycB7ZltXRc1yPcCX6FjwX8Bidqu79Ezxt8S+AOELpKu7Iiy9vXGd/ZpEUiz83qw88eKAsT/l/j6gwP7s9y2Y8hNZMS+3XhhL4aZ0hwhhHHu4+3XDVx4ybm8wdJ+95eMFeaMYG9eqMXJKZ68o+wgfNmItuy+pByDDtpWY4+9ne/Y8J0InKRzMTjjZgZwg1E8f0Ge7zCUP9nXTkTpjW+5yB4OjtwZoxXKkxFWLjfr7kMs5l3bLJ/AwAA//+ElzrnUgkAAA==",
	} {
		charts[file] = decode(file, encoded)
	}
}
