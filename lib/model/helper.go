/*
 * Copyright 2022 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

import (
	"github.com/SENERGY-Platform/models/go/models"
	"strings"
)

func SetFunctionRdfType(function *models.Function) {
	if strings.HasPrefix(function.Id, URN_PREFIX+"controlling-function:") {
		function.RdfType = SES_ONTOLOGY_CONTROLLING_FUNCTION
	}
	if strings.HasPrefix(function.Id, URN_PREFIX+"measuring-function:") {
		function.RdfType = SES_ONTOLOGY_MEASURING_FUNCTION
	}
}

// ConceptFunctionRdfTypes are the function types a concept generates one function for.
var ConceptFunctionRdfTypes = []string{SES_ONTOLOGY_MEASURING_FUNCTION, SES_ONTOLOGY_CONTROLLING_FUNCTION}

// ConceptFunctionName renders the name of the function a concept generates for a function
// type. A concept creates both when it is created and renames them when it is renamed, so
// that every concept carries one obvious pair. Returns "" for anything that is not one of
// the two function types.
func ConceptFunctionName(rdfType string, conceptName string) string {
	switch rdfType {
	case SES_ONTOLOGY_MEASURING_FUNCTION:
		return "Get-" + conceptName
	case SES_ONTOLOGY_CONTROLLING_FUNCTION:
		return "Set-" + conceptName
	}
	return ""
}
