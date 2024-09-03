<a href="https://github.com/SENERGY-Platform/device-repository/actions/workflows/tests.yml" rel="nofollow">
    <img src="https://github.com/SENERGY-Platform/device-repository/actions/workflows/tests.yml/badge.svg?branch=master" alt="Tests" />
</a>

## OpenAPI
uses https://github.com/swaggo/swag

### generating
```
go generate ./...
```

### swagger ui
if the config variable UseSwaggerEndpoints is set to true, a swagger ui is accessible on /swagger/index.html (http://localhost:8080/swagger/index.html)