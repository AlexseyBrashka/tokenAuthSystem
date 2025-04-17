Точки входа( через docker-compose up):

0.0.0.0:8080/auth/tokens - получить токен
Пример: контента для post запроса: {"user_id": "123e4567-e89b-12d3-a456-426614174000"}

0.0.0.0:8080/auth/refresh - обновить токен
Пример: контента для post запроса: {"token":"ZjQyYTk3YjEtMmE4Zi00MTIwLTk5OGUtODVhYmE5ZjAxNDAyOl9DMFhRRzBxdXlrVmxLMzJjZnk0cFdnelcyYWdQUFp0ZS1scUU5dUMxVkk9",
"token_type":"RefreshToken"} -( при валидном токине )

token-type - позволяет обновлять как по RefreshToken, так и по AccessToken.
При этом: Access токен при активации - блокирует, как свое повторное использование, так и использование выданного с ним RefreshToken
