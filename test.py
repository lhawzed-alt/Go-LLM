from openai import OpenAI

client = OpenAI(
    api_key="sk-relay-test-key-123",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="agnes-2.5-flash",
    messages=[
        {"role": "user", "content": "你好"}
    ]
)

print(response.choices[0].message.content)