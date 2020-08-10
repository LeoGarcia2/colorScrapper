package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func separator() {
	fmt.Println("\n-----------------------\n")
}

func trimDuplicate(arrayStrings []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range arrayStrings {
		if _, value := keys[strings.ToLower(entry)]; !value {
			keys[entry] = true
			list = append(list, strings.ToLower(entry))
		}
	}
	return list
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Write the URL from which you want to pull the colors :")

	for {
		fmt.Print("-> ")

		colorsFound := []string{}

		url, _ := reader.ReadString('\n')
		url = strings.Replace(url, "\n", "", -1)
		url = strings.Replace(url, "\r", "", -1)

		response, error := http.Get(url)

		separator()

		if error != nil {
			fmt.Println("Error message :")
			fmt.Println(error)
		}

		if response != nil {
			defer response.Body.Close()

			if response.StatusCode == http.StatusOK {
				bodyBytes, error := ioutil.ReadAll(response.Body)
				if error != nil {
					fmt.Println(error)
				}

				html := string(bodyBytes)

				regexColors := regexp.MustCompile("(#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})|rgb(.)*?[0-9]+\\))")
				regexStyle := regexp.MustCompile("<link (.)*?rel=\"stylesheet\"(.)*?>")

				matchedColors := regexColors.FindAllString(html, -1)
				matchedStyle := regexStyle.FindAllString(html, -1)

				colorsFound = append(colorsFound, matchedColors...)

				fmt.Println("Colors found in html :")

				separator()

				fmt.Printf("%q\n", matchedColors)

				separator()

				fmt.Println("Style sources found :")

				for i := 0; i < len(matchedStyle); i++ {
					regexHref := regexp.MustCompile("href=\"(.*?)\"")
					matchedHref := regexHref.FindStringSubmatch(matchedStyle[i])

					isRelativeDotted := strings.HasPrefix(matchedHref[1], ".")
					// fmt.Println("'./' : ", isRelativeDotted)
					isAbsolute := strings.HasPrefix(matchedHref[1], "http")
					// fmt.Println("'abs' : ", isAbsolute)

					linkToStylesheet := ""

					if isAbsolute {
						linkToStylesheet = matchedHref[1]
					} else {
						if strings.HasSuffix(url, "/") && strings.HasPrefix(matchedHref[1], "/") {
							url = strings.TrimRight(url, "/")
						}
						if isRelativeDotted {
							linkToStylesheet = url + strings.TrimLeft(matchedHref[1], ".")
						} else {
							if !strings.HasSuffix(url, "/") && !strings.HasPrefix(matchedHref[1], "/") {
								url = url + "/"
							}
							linkToStylesheet = url + matchedHref[1]
						}
					}

					separator()

					fmt.Println("Absolute stylesheet url : ", linkToStylesheet)

					response, error := http.Get(linkToStylesheet)
					if error != nil {
						fmt.Println("Error message :")
						fmt.Println(error)
					}

					if response != nil {
						defer response.Body.Close()

						if response.StatusCode == http.StatusOK {
							styleBytes, error := ioutil.ReadAll(response.Body)
							if error != nil {
								fmt.Println(error)
							}

							css := string(styleBytes)

							matchedColors := regexColors.FindAllString(css, -1)

							colorsFound = append(colorsFound, matchedColors...)

							fmt.Println("Colors found in css :")

							fmt.Printf("%q\n", matchedColors)
						}
					}
				}

				separator()

				colorsFound = trimDuplicate(colorsFound)
				colorsFound = trimDuplicate(colorsFound)

				fmt.Println("All colors found :")
				fmt.Println(colorsFound)

				file, error := os.Create("colors.css")
				if error != nil {
					fmt.Println(error)
					file.Close()
					return
				}

				fmt.Fprintln(file, "* {")

				for _, line := range colorsFound {
					fmt.Fprintln(file, "\tcolor: "+line+";")
				}

				fmt.Fprintln(file, "}")

				error = file.Close()
				if error != nil {
					fmt.Println(error)
					return
				}

				fmt.Println("Colors saved into colors.css !")
			}
		}
	}
}
