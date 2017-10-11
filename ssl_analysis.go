package main

import (
	"net/http"
	"fmt"
	"crypto/tls"
	"os"
	"crypto/x509"
	"regexp"
	"strings"
)

func main() {
	if strings.HasPrefix(string(os.Args[1]), "https://") {
		test_ssl(os.Args[1])
	} else if strings.EqualFold(string(os.Args[1]), "options") {
		fmt.Println("HSTS: Enforce the use of TLS/SSL in an user agent")
		fmt.Println("Content Security Policy: Helpful to protect your site against XSS attacks")
		fmt.Println("X-Frame-Options: Preventing a browser from framing your site. Helpful against clickjacking")
		fmt.Println("X-XSS-Protection: Configure XSS Protection in Chrome, Safari and IE")
		fmt.Println("X-Content-Type-Options: Stops Browser from Sniffing the content type")
		fmt.Println("Referer-Policy: Allow the site to control the value of the referer header in links away " +
			"from their pages")
	} else {
		fmt.Println("Usage: go run ssl_analysis.go https://www.google.de")
		fmt.Println("If you want to get information about the security options in the header use: \n " +
			"go run ssl_analysis.go options")
	}
}

func test_ssl(domain string){
	data, err := http.Get(domain)
	if err != nil {
		fmt.Println("There seems to be a problem with the certificate of", domain)
		fmt.Println(err)
		fmt.Println("Trying with skipped Security Verification...")
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		data2, err := client.Get(domain)
		if err != nil {
			fmt.Println(err)
		}else{
			fmt.Println("Skip Security Verification..")
			temp_header := data2.Header
			print_headers(temp_header)
			var_temp := domain[8:] + ":443"
			conf := &tls.Config{
				InsecureSkipVerify: true,
			}
			conn, err := tls.Dial("tcp", var_temp, conf)
			if err != nil {
				fmt.Println(err)
			}
			cert := conn.ConnectionState().PeerCertificates[0]

			defer conn.Close()
			print_values(cert)
		}
	} else {
		fmt.Println("Certificate seems okay... Lets check the HTTPS Response Header... \n ")
		temp_header := data.Header
		print_headers(temp_header)
		fmt.Println(" \nChecking the certificate...")
		var_temp := domain[8:] + ":443"
		conn, err := tls.Dial("tcp", var_temp, nil)
		if err != nil {
			fmt.Println(err)
		}
		cert := conn.ConnectionState().PeerCertificates[0]
		defer conn.Close()
		print_values(cert)
	}
}

func print_values(cert *x509.Certificate) {
	fmt.Println("\nThe Certificate was Issued by:\n", cert.Issuer)
	fmt.Println("Here are some additional Information about the Certificate")
	fmt.Println("Subject:", cert.Subject)
	fmt.Println("Starts:", cert.NotBefore)
	fmt.Println("Expires:", cert.NotAfter)
	fmt.Println("DNS Names:", cert.DNSNames)
	fmt.Println("Crypto-Algorithm:  ", cert.SignatureAlgorithm)
	fmt.Println("Issues URL:  ", cert.IssuingCertificateURL)
}

func print_headers(temp_header http.Header) {
	arr_regex := [7]string{ "X-Xss-Protection", "X-Frame-Options", "Strict-Transport-Security",
		"Content-Security-Policy", "X-Content-Type-Options", "Public-Key-Pins", "Referrer-Policy" }
	fmt.Println("Your HTTPS Response was checked for these Security Options: ", arr_regex)
	x := 0
	y := 0
	srv_vers := "nil"
	for _, value := range arr_regex {
		for key, val := range temp_header {
			if key == "Server" {
				justString := strings.Join(val," ")
				srv_vers = justString
			}
			r, err := regexp.Compile(value)
			if err != nil {
				fmt.Printf("There is a problem with your regexp.\n")
				continue
			}
			if r.MatchString(key) == true {
				fmt.Println("The following Security Option was found: ", key)
				x += 1
				y = 1
			}
		}
	if y==0 {
		fmt.Println("NOT implemented ", value)
	}else {
		y = 0
	}
	}
	fmt.Println(x, " activated Security Options were found in your response!")
	if x == 0 {
		fmt.Println("You have no Security Options activated! You should do this immediately!")
	}
	fmt.Println("\n\nServer Version is:",srv_vers)


}