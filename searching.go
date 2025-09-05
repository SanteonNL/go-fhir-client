/*
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package fhirclient

import (
	"context"
	"fmt"
	"net/url"

	"github.com/zorgbijjou/golang-fhir-models/fhir-models/fhir"
)

// Paginate is a utility function to scan through all pages of a FHIR search result.
// It calls the consumeFunc for each page, which can return false to stop the pagination early (if for example enough data has been found).
// The function will stop if there are no more pages (no "next" link in the Bundle).
// It will return an error if any of the calls to consumeFunc or the FHIR server fail.
// By default, it will stop after 100 iterations to prevent endless loops due to bugs in the FHIR server or the code.
func Paginate(ctx context.Context, fhirClient Client, searchSet fhir.Bundle, consumeFunc func(*fhir.Bundle) (bool, error), opts ...PaginationOption) error {
	options := &paginationOptions{
		maxIterations: 100,
	}
	for _, opt := range opts {
		opt(options)
	}
	var nextURL *url.URL
	for i := 0; i < options.maxIterations; i++ {
		// Make sure we don't loop endlessly due to a bug
		if i == options.maxIterations-1 {
			return fmt.Errorf("paginate: max. search iterations reached (%d), possible bug", options.maxIterations)
		}

		if proceed, err := consumeFunc(&searchSet); err != nil {
			return err
		} else if !proceed {
			// consume function called exit
			return nil
		}

		hasNext := false
		for _, link := range searchSet.Link {
			if link.Relation == "next" {
				var err error
				if nextURL, err = url.Parse(link.Url); err != nil {
					return fmt.Errorf("paginate: invalid 'next' link for search set: %w", err)
				}
				hasNext = true
			}
		}
		if !hasNext {
			break
		}
		searchSet = fhir.Bundle{}
		if err := fhirClient.SearchWithContext(ctx, "", nil, &searchSet, AtUrl(nextURL)); err != nil {
			return fmt.Errorf("pagintate: query next page failed (url=%s): %w", nextURL, err)
		}
	}
	return nil
}

type PaginationOption func(*paginationOptions)

type paginationOptions struct {
	maxIterations int
}

// WithMaxIterations sets the maximum number of iterations for the Paginate function.
func WithMaxIterations(max int) PaginationOption {
	return func(o *paginationOptions) {
		o.maxIterations = max
	}
}
