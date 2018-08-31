// Copyright 2018 Google LLC

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     https://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicemanagement/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Config is the configuration structure used by the LambdaController
type Config struct {
	Project          string
	ProjectNum       string
	clientCompute    *compute.Service
	clientServiceMan *servicemanagement.APIService
	clientset        *kubernetes.Clientset
	serviceAccount   string
}

func (c *Config) loadAndValidate() error {
	var err error

	if c.Project == "" {
		log.Printf("[INFO] Fetching Project ID from Compute metadata API...")
		c.Project, err = metadata.ProjectID()
		if err != nil {
			return err
		}
	}

	if c.ProjectNum == "" {
		log.Printf("[INFO] Fetching Numeric Project ID from Compute metadata API...")
		c.ProjectNum, err = metadata.NumericProjectID()
		if err != nil {
			return err
		}
	}

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return err
	}
	c.clientset = clientset

	clientScopes := []string{
		compute.ComputeScope,
		servicemanagement.ServiceManagementScope,
	}

	client, err := google.DefaultClient(oauth2.NoContext, strings.Join(clientScopes, " "))
	if err != nil {
		return err
	}

	log.Printf("[INFO] Instantiating GCE client...")
	c.clientCompute, err = compute.New(client)

	log.Printf("[INFO] Instantiating Google Cloud Service Management Client...")
	c.clientServiceMan, err = servicemanagement.New(client)
	if err != nil {
		return err
	}

	return nil
}
