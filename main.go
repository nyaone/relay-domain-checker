package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	RelayListDomainsFileName = "domains.txt"
	ResultSaveFileName       = "domains.json"
	RequestTimeout           = 10 * time.Second
)

func main() {
	domains, err := os.ReadFile(RelayListDomainsFileName)
	if err != nil {
		log.Fatal(err)
	}

	fileLines := strings.Split(strings.ReplaceAll(string(domains), "\r", ""), "\n")

	totalDomainCounter := 0

	var unresolvedList []string
	var notFunctioningList []string
	var wrongCodeList []domainWithErrorCode

	var misformattedNodeInfoListList []string
	var noAvailableNodeInfoSchemaList []string
	var misformattedNodeInfoSchemaList []string
	var validList []domainWithValidNodeinfo

	var wg sync.WaitGroup

	for _, line := range fileLines {

		if strings.Contains(line, "[*]") {
			// Is domain
			wg.Add(1)

			totalDomainCounter++

			domain := strings.Replace(line, "[*] ", "", 1)

			go func() {
				defer wg.Done()

				// Try resolve
				_, err := net.LookupIP(strings.Split(domain, ":")[0])
				if err != nil {
					unresolvedList = append(unresolvedList, domain)
					log.Printf("[%s] is unresolved.", domain)
					return
				}

				// Try get nodeinfo
				wellknownNodeinfo := fmt.Sprintf("https://%s/.well-known/nodeinfo", domain)

				resp, err := (&http.Client{
					Timeout: RequestTimeout,
				}).Get(wellknownNodeinfo)
				if err != nil {
					notFunctioningList = append(notFunctioningList, domain)
					log.Printf("[%s] is not functioning with error: %v.", domain, err)
					return
				} else if resp.StatusCode != http.StatusOK {
					wrongCodeList = append(wrongCodeList, domainWithErrorCode{
						Domain: domain,
						Code:   resp.StatusCode,
					})
					log.Printf("[%s] is [RETURNING %d].", domain, resp.StatusCode)
					return
				}

				log.Printf("[%s] is valid, gathering nodeinfo.", domain)

				// Bind nodeinfo list to json
				var infoList NodeInfoList
				err = json.NewDecoder(resp.Body).Decode(&infoList)
				if err != nil {
					misformattedNodeInfoListList = append(misformattedNodeInfoListList, domain)
					log.Printf("[%s] 's nodeinfo list is invalid with error: %v.", domain, err)
					return
				}

				// Get nodeinfo
				nodeinfoSchemaLink := ""
				for _, item := range infoList.Links {
					if strings.HasPrefix(item.Rel, "http://nodeinfo.diaspora.software/ns/schema/2.") {
						// Both 2.0 and 2.1 is fine
						nodeinfoSchemaLink = item.Href
						break
					}
				}

				if nodeinfoSchemaLink == "" {
					// No valid nodeinfo schema
					noAvailableNodeInfoSchemaList = append(noAvailableNodeInfoSchemaList, domain)
					log.Printf("[%s] doesn't have an available nodeinfo schema.", domain)
					return
				}

				resp, err = (&http.Client{
					Timeout: RequestTimeout,
				}).Get(nodeinfoSchemaLink)
				if err != nil {
					noAvailableNodeInfoSchemaList = append(noAvailableNodeInfoSchemaList, domain)
					log.Printf("[%s] 's schema href is not accessible with error: %v.", domain, err)
					return
				}

				var infoSchema NodeInfoSchema
				err = json.NewDecoder(resp.Body).Decode(&infoSchema)
				if err != nil {
					misformattedNodeInfoSchemaList = append(misformattedNodeInfoSchemaList, domain)
					log.Printf("[%s] 's nodeinfo schema is invalid with error: %v.", domain, err)
					return
				}

				// Finally this instance can be marked as valid
				validList = append(validList, domainWithValidNodeinfo{
					Domain:   domain,
					NodeInfo: infoSchema,
				})
			}()
		}
	}

	wg.Wait()

	log.Printf(
		"Found %d unresolved, %d not functioning, %d wrong code, %d is still valid",
		len(unresolvedList),
		len(notFunctioningList),
		len(wrongCodeList),
		len(misformattedNodeInfoListList)+len(noAvailableNodeInfoSchemaList)+len(misformattedNodeInfoSchemaList)+len(validList),
	)

	// Read last time result
	var lastTimeResult ResultFileFormat

	if _, err = os.Stat(ResultSaveFileName); err != nil {
		log.Printf("Failed to check last time result file with error: %v.", err)
	} else if lastTimeResultBytes, err := os.ReadFile(ResultSaveFileName); err != nil {
		log.Printf("Failed to read last time result file with error: %v.", err)
	} else if err = json.Unmarshal(lastTimeResultBytes, &lastTimeResult); err != nil {
		log.Printf("Failed to unmarshal last time result file with error: %v.", err)
	}

	// Generate current result file (json)
	currentTime := time.Now()
	var result ResultFileFormat
	result.CollectedAt = currentTime

	result.Unresolved = InheritStatus(lastTimeResult.Unresolved, unresolvedList, currentTime)
	result.NotFunctioning = InheritStatus(lastTimeResult.NotFunctioning, notFunctioningList, currentTime)
	result.WrongCode = InheritStatusAndCode(lastTimeResult.WrongCode, wrongCodeList, currentTime)

	result.MisformattedNodeInfoList = InheritStatus(lastTimeResult.MisformattedNodeInfoList, misformattedNodeInfoListList, currentTime)
	result.NoAvailableNodeInfoSchema = InheritStatus(lastTimeResult.NoAvailableNodeInfoSchema, noAvailableNodeInfoSchemaList, currentTime)
	result.MisformattedNodeInfoSchema = InheritStatus(lastTimeResult.MisformattedNodeInfoSchema, misformattedNodeInfoSchemaList, currentTime)

	validResults := make(ResultValidWithNodeInfo)
	for _, validDomain := range validList {
		validResults[validDomain.Domain] = validDomain.NodeInfo
	}
	result.Valid = validResults

	resultBytes, err := json.Marshal(&result)
	if err != nil {
		log.Fatalf("Failed to format result into bytes with error: %v", err)
		return
	}

	err = os.WriteFile(ResultSaveFileName, resultBytes, 0644)
	if err != nil {
		log.Fatalf("Failed to save result into file with error: %v", err)
	}

	log.Printf("%d domains validate finished successfully.", totalDomainCounter)

}

func InheritStatus(oldList ResultErrRecord, currentList []string, currentTime time.Time) ResultErrRecord {
	newList := make(ResultErrRecord)
	for _, errDomain := range currentList {
		dErr := currentTime
		// Find if it's first time or has been a while
		if len(oldList) > 0 {
			for oldErrDomain, oldErrSince := range oldList {
				if oldErrDomain == errDomain {
					newList[errDomain] = oldErrSince
					break
				}
			}
		}
		newList[errDomain] = dErr
	}
	return newList
}

func InheritStatusAndCode(oldList ResultErrRecordWithCode, currentList []domainWithErrorCode, currentTime time.Time) ResultErrRecordWithCode {
	newList := make(ResultErrRecordWithCode)
	for _, errDomain := range currentList {
		dErr := ErrorStatusWithCode{
			Code:  errDomain.Code,
			Since: currentTime,
		}
		// Find if it's first time or has been a while
		if len(oldList) > 0 {
			for oldErrDomain, oldErrRecord := range oldList {
				if oldErrDomain == errDomain.Domain && oldErrRecord.Code == errDomain.Code {
					dErr.Since = oldErrRecord.Since
					break
				}
			}
		}
		newList[errDomain.Domain] = dErr
	}
	return newList
}
