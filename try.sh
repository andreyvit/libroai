curl https://api.openai.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "system",
      "content": "Make SHORT chat title max 5 words"
    },
    {
      "role": "user",
      "content": "Ive seen great wins in my first month in Tribe, but is this just a honeymoon phase? How do I keep this momentum?"
    }
  ],
  "functions": [
    {
      "name": "set_chat_title",
      "description": "Set the title of the chat.",
      "parameters": {
        "type": "object",
        "required": [
          "title",
          "subtitle"
        ],
        "properties": {
          "title": {
            "type": "string",
            "description": "5 words or less"
          }
        }
      }
    }
  ],
  "function_call": {"name": "set_chat_title"},
  "temperature": 0.5,
  "max_tokens": 50,
  "top_p": 1,
  "frequency_penalty": 0,
  "presence_penalty": 0
}'
