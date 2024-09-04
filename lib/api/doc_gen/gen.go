/*
 * Copyright 2024 InfAI (CC SES)
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

package main

import "github.com/SENERGY-Platform/permissions-v2/pkg/client"

//go:generate go run gen.go
//go:generate go install github.com/swaggo/swag/cmd/swag@latest
//go:generate swag init --instanceName devicerepository -o ../../../docs --parseDependency -d .. -g api.go

// generates lib/api/permissions.go
// which enables swag init to generate documentation for permissions endpoints
// which are added by 'permForward := client.New(config.PermissionsV2Url).EmbedPermissionsRequestForwarding("/permissions/", router)'
func main() {
	err := client.GenerateGoFileWithSwaggoCommentsForEmbededPermissionsClient("api", "permissions", "../generated_permissions.go")
	if err != nil {
		panic(err)
	}
}