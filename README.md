<div align="center">
  <img src="https://u.julep.ai/responses-logo.png" alt="Open Responses Logo" width="200"/>
  <h1>Julep Open Responses</h1>
  <p><i>Self-hosted, open-source alternative to OpenAI's Responses API</i></p>
</div>

<div align="center">
  <a href="https://docs.julep.ai/responses/quickstart">Documentation</a> •
  <a href="https://github.com/julep-ai/julep">GitHub</a> •
  <a href="https://docs.julep.ai/responses/examples">Examples</a>
</div>

<br>

## 💡 Overview

Julep's Open Responses lets you generate content with any LLM backend without creating persistent agents or sessions. It's a drop-in replacement for OpenAI's Responses API that you control.

## ✨ Features

- **🔄 Model Flexibility** - Use any LLM backend (Claude, Qwen, Deepseek R1, and more)
- **🔒 Self-hosted & Private** - Full control over your deployment
- **🔌 Drop-in Compatible** - Works with official Agents SDK by just changing the URL
- **🚀 Easy Setup** - Deploy in minutes with Docker or our CLI
- **🛠️ Built-in Tools** - Automatic tool calls like web_search using open alternatives

## 🚀 Quick Start

One command and you're ready to go:

```bash
npx -y open-responses init
```

## 📦 Installation Options

<details>
<summary><b>Docker Installation</b></summary>

```bash
mkdir julep-responses-api
cd julep-responses-api
wget https://u.julep.ai/responses-env.example -O .env
wget https://u.julep.ai/responses-compose.yaml -O docker-compose.yml
docker compose up --watch
```
</details>

<details>
<summary><b>CLI Installation</b></summary>

```bash
# Using npx
npx open-responses setup
npx open-responses start

# Using uv
uvx open-responses setup
uvx open-responses start
```
</details>

## 💻 Usage Examples

<details>
<summary><b>Node.js</b></summary>

```javascript
import { OpenAI } from 'openai';

const client = new OpenAI({ 
  baseURL: 'http://localhost:8080/', 
  apiKey: "RESPONSE_API_KEY" 
});

const response = await client.responses.create({
  model: "gpt-4o-mini",
  input: "How many people live in the world?"
});

console.log(response.output[0].content[0].text);
```
</details>

<details>
<summary><b>Python</b></summary>

```python
import os
from openai import OpenAI

client = OpenAI(
  base_url="http://localhost:8080/", 
  api_key=os.getenv("RESPONSE_API_KEY")
)

response = client.responses.create(
  model="gpt-4o-mini",
  input="How many people live in the world?"
)

print(response.output[0].content[0].text)
```
</details>

## 📚 Documentation

- [🚀 Quickstart Guide](https://docs.julep.ai/responses/quickstart)
- [🧠 API Concepts](https://docs.julep.ai/responses/concepts)
- [📝 Code Examples](https://docs.julep.ai/responses/examples)
- [💻 CLI Documentation](https://docs.julep.ai/responses/cli)
- [🔮 Project Roadmap](https://docs.julep.ai/responses/roadmap)

## 🤝 Contributing

Licensed under Apache-2.0. We welcome your feedback and contributions!

[Visit our GitHub repository](https://github.com/julep-ai/julep)
