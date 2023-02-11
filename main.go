package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

func main() {
	domains, err := os.ReadFile("domains.txt")
	if err != nil {
		log.Fatal(err)
	}

	domainsList := strings.Split(strings.Replace(string(domains), "\r", "", -1), "\n")

	var unresolvedDomains []string
	var notFunctioningDomains []string
	var wrongcodeDomains []string
	var validDomains []string

	var wg sync.WaitGroup

	for _, domain := range domainsList {

		if !strings.Contains(domain, " ") && strings.Contains(domain, ".") {
			// Is domain
			wg.Add(1)

			domain := domain

			go func() {
				defer wg.Done()

				// Try resolve
				_, err := net.LookupIP(strings.Split(domain, ":")[0])
				if err != nil {
					unresolvedDomains = append(unresolvedDomains, domain)
					log.Printf("[%s] is [UNRESOLVED].", domain)
					return
				}

				// Try get nodeinfo
				wellknownNodeinfo := fmt.Sprintf("https://%s/.well-known/nodeinfo", domain)
				resp, err := http.Get(wellknownNodeinfo)
				if err != nil {
					notFunctioningDomains = append(notFunctioningDomains, domain)
					log.Printf("[%s] is [NOT FUNCTIONING].", domain)
					return
				} else if resp.StatusCode != http.StatusOK {
					wrongcodeDomains = append(wrongcodeDomains, domain)
					log.Printf("[%s] is [RETURNING %d].", domain, resp.StatusCode)
					return
				}

				// Mark as valid
				validDomains = append(validDomains, domain)
				log.Printf("[%s] is valid.", domain)
			}()
		}
	}

	wg.Wait()

	log.Printf(
		"Found %d unresolved, %d not functioning, %d wrong code, %d is still valid",
		len(unresolvedDomains),
		len(notFunctioningDomains),
		len(wrongcodeDomains),
		len(validDomains),
	)

	os.WriteFile("unresolved.txt", []byte(strings.Join(unresolvedDomains, "\n")), 0644)
	os.WriteFile("notfunctioning.txt", []byte(strings.Join(notFunctioningDomains, "\n")), 0644)
	os.WriteFile("wrongcode.txt", []byte(strings.Join(wrongcodeDomains, "\n")), 0644)
	os.WriteFile("valid.txt", []byte(strings.Join(validDomains, "\n")), 0644)

}
