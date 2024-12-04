package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-autoinst/gopenqa"
)

/* Read products from stdin */
func readProducts(filename string) ([]gopenqa.Product, error) {
	var data []byte
	var err error

	if filename == "" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return make([]gopenqa.Product, 0), err
		}
	} else {
		// TODO: Don't use io.ReadAll
		if file, err := os.Open(filename); err != nil {
			return make([]gopenqa.Product, 0), err
		} else {
			defer file.Close()
			data, err = io.ReadAll(file)
			if err != nil {
				return make([]gopenqa.Product, 0), err
			}
		}
	}

	// First try to read a single product
	var product gopenqa.Product
	if err := json.Unmarshal(data, &product); err == nil {
		products := make([]gopenqa.Product, 0)
		products = append(products, product)
		return products, nil
	}

	// Then try to read a product array
	var products []gopenqa.Product
	if err := json.Unmarshal(data, &products); err == nil {
		return products, err
	}

	products = make([]gopenqa.Product, 0)
	return products, fmt.Errorf("invalid input format")
}

func postProduct(args []string) error {
	files := args
	if len(files) == 0 {
		files = append(files, "")
	}

	for _, filename := range files {
		if products, err := readProducts(filename); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		} else {
			for _, product := range products {
				if product, err := instance.PostProduct(product); err != nil {
					return err
				} else {
					fmt.Printf("Posted product %d\n", product.ID)
				}
			}
		}
	}
	return nil
}

func runProducts(args []string) error {
	method := "GET"

	if len(args) > 0 {
		// get method
		method = args[0]
		args = args[1:]
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "GET" {
		products, err := instance.GetProducts()
		if err != nil {
			return err
		}
		if err := printJson(products); err != nil {
			return err
		}
		return nil
	} else if method == "POST" {
		return postProduct(args)
	} else {
		return fmt.Errorf("invalid method: %s", method)
	}
}

func runProduct(args []string) error {
	method := "GET"

	if len(args) > 0 {
		// get method
		method = args[0]
		args = args[1:]
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "GET" {
		ids, _ := extractIntegers(args)
		if len(ids) == 0 {
			return fmt.Errorf("missing product ids")
		}
		for _, id := range ids {
			product, err := instance.GetProduct(id)
			if err != nil {
				return err
			}

			if err := printJson(product); err != nil {
				return err
			}
		}
		return nil
	} else if method == "POST" {
		return postProduct(args)
	} else {
		return fmt.Errorf("invalid method: %s", method)
	}
}
