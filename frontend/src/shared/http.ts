export async function request<T>(
  path: string,
  method = "GET",
  body?: unknown,
): Promise<T> {
  let response: Response;
  try {
    response = await fetch(path, {
      method,
      headers:
        body === undefined ? undefined : { "Content-Type": "application/json" },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new Error(
      "Не удалось связаться с сервером. Проверьте подключение и попробуйте ещё раз.",
    );
  }

  if (!response.ok) {
    let message = `Сервер вернул ошибку ${response.status}.`;
    if (
      response.headers.get("Content-Type")?.includes("application/problem+json")
    ) {
      try {
        const problem: unknown = await response.json();
        if (
          problem &&
          typeof problem === "object" &&
          "detail" in problem &&
          typeof problem.detail === "string"
        ) {
          message = problem.detail;
        }
      } catch {
        // Если тело ответа повреждено, показываем HTTP-статус.
      }
    }
    throw new Error(message);
  }

  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}
