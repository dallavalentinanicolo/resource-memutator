package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Main handler for the webhook
// HandlerFunc need a function takes 2 parameters for
// ServeHTTP(w ResponseWriter, r *Request)
func mutate(w http.ResponseWriter, r *http.Request) {
	// Read the incoming request
	var (
		body []byte
		err  error
	)
	// ReadAll read a string the string is in r.Body ( r request)
	// request has been done and received the body
	body, err = io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Couldn't read request body", http.StatusBadRequest)
		return
	}

	// Parse the AdmissionReview request
	var admissionReview admissionv1.AdmissionReview
	// deserialize a Json in a struc
	// input body, output admissionReview
	// just return an error if exist
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		http.Error(w, "Couldn't parse request", http.StatusBadRequest)
		return
	}

	// Create a response object
	// create a struc of responde
	var response admissionv1.AdmissionResponse = admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true, // Allow the request by default
	}

	// Check if the object is a Deployment

	switch admissionReview.Request.Kind.Kind {
	case "Deployment":
		var deployment appsv1.Deployment
		if err := json.Unmarshal(admissionReview.Request.Object.Raw, &deployment); err != nil {
			http.Error(w, "Couldn't parse deployment", http.StatusBadRequest)
			return
		}
		checkMemoryLimits(deployment.Spec.Template.Spec.Containers, deployment.Name, "Deployment")
	case "StatefulSet":
		var statefulset appsv1.StatefulSet
		if err := json.Unmarshal(admissionReview.Request.Object.Raw, &statefulset); err != nil {
			http.Error(w, "Couldn't parse StatefulSet", http.StatusBadRequest)
			return
		}
		checkMemoryLimits(statefulset.Spec.Template.Spec.Containers, statefulset.Name, "StatefulSet")
	}

	// Wrap the response into an AdmissionReview
	admissionReview.Response = &response
	respBytes, _ := json.Marshal(admissionReview)
	w.Write(respBytes)
}

// checkMemoryLimits checks if memory requests and limits match for containers and logs mismatches
func checkMemoryLimits(containers []corev1.Container, resourceName, resourceType string) {
	for _, container := range containers {
		requestMem := container.Resources.Requests[corev1.ResourceMemory]
		limitMem := container.Resources.Limits[corev1.ResourceMemory]

		// Log if requests and limits differ
		if requestMem.Cmp(limitMem) != 0 {
			log.Printf("Container '%s' in %s '%s' has mismatched memory requests and limits. Request: %s, Limit: %s\n",
				container.Name, resourceType, resourceName, requestMem.String(), limitMem.String())
		}
	}
}

func main() {
	var (
		// LoadX509KeyPair (tls.Certificate, error)
		cert   tls.Certificate
		err    error
		server *http.Server
	)
	// TODO improve the certification generation maybe Cert manager
	// func of library TLS then tls contains a library LoadX509KeyPair
	cert, err = tls.LoadX509KeyPair("/Certificate/tls.crt", "/Certificate/tls.key")
	if err != nil {
		log.Printf("Error loading key pair: %v\n", err)
		os.Exit(1)
	}

	// Setup the parameters for running HTTP server for the webhook
	server = &http.Server{
		Addr:    ":443",
		Handler: http.HandlerFunc(mutate), // Your mutate handler
		TLSConfig: &tls.Config{
			// generate a certificate based on tls.LoadX509KeyPair("/Certificate/tls.crt", "/Certificate/tls.key")
			Certificates: []tls.Certificate{cert},
		},
	}

	// Start serving HTTPS
	log.Printf("Starting webhook server on port 443...")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Printf("Error starting server: %v\n", err)
	}
}
