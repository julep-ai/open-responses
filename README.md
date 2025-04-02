<div align="center">
  <img src="https://github.com/user-attachments/assets/0a216411-e689-4b08-97f9-7e3e01ce5afc" alt="Open Responses Logo" width="320"/>
  <img src="https://github.com/user-attachments/assets/21cff538-e5b8-4c89-b19f-3280f65b2105" alt="Launching Open Responses" width="320"/>
  <h1>Open Responses</h1>
  <p><i>Self-hosted, open-source alternative to OpenAI's Responses API</i></p>
</div>

<div align="center">
  <a href="https://docs.julep.ai/responses/quickstart">Documentation</a> •
  <a href="https://github.com/julep-ai/julep">GitHub</a> •
  <a href="https://docs.julep.ai/responses/examples">Examples</a>
</div>

<br>

## 💡 What is Open Responses?

Open Responses lets you run a fully self-hosted version of OpenAI's Responses API. It works seamlessly with any large language model (LLM) provider—whether it's Claude, Qwen, Deepseek R1, Ollama, or others. Swap out OpenAI without changing your existing Agents SDK integrations.

> [!NOTE]
> This project is developed by the team behind **[Julep AI](https://julep.ai)**, the open-source platform making it easy for data teams to build, deploy, and scale stateful AI agents and workflows. Check us out on github: [![GitHub Repo stars](https://img.shields.io/github/stars/julep-ai/julep?style=social&label=julep-ai%2Fjulep&link=https%3A%2F%2Fgithub.com%2Fjulep-ai%2Fjulep)
](https://github.com/julep-ai/julep)


## ✨ Why use Open Responses?

- **🔄 Bring Your Own Model** - Compatible with any LLM provider you prefer.
- **🔒 Privacy First** - Fully self-hosted, giving you total control over your data.
- **🔌 Easy Switch** - Drop-in replacement compatible with OpenAI’s official Agents SDK.
- **🚀 Fast Setup** - Get started quickly with Docker or our straightforward CLI.
- **🛠️ Built-in Tools** - Supports automatic tool calls like web searches using open-source alternatives.

## 🚀 Quick Start

One simple command to get going:

```bash
npx -y open-responses init
```

## 📦 Installation Methods

```bash
mkdir julep-responses-api
cd julep-responses-api
wget https://u.julep.ai/responses-env.example -O .env
wget https://u.julep.ai/responses-compose.yaml -O docker-compose.yml
docker compose up --watch
```

```bash
# Using npx
npx open-responses setup
npx open-responses start

# Using uv
uvx open-responses setup
uvx open-responses start
```

## 💻 Quick Examples

```javascript
import { OpenAI } from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:8080/',
  apiKey: "RESPONSE_API_KEY"
});

const response = await client.responses.create({
  model: "gpt-4o-mini",
  input: "What's the population of the world today?"
});

console.log(response.output[0].content[0].text);
```

```python
import os
from openai import OpenAI

client = OpenAI(
  base_url="http://localhost:8080/",
  api_key=os.getenv("RESPONSE_API_KEY")
)

response = client.responses.create(
  model="gpt-4o-mini",
  input="What's the population of the world today?"
)

print(response.output[0].content[0].text)
```

## 📚 Learn More

- [🚀 Quickstart Guide](https://docs.julep.ai/responses/quickstart)
- [🧠 API Concepts](https://docs.julep.ai/responses/concepts)
- [📝 Code Examples](https://docs.julep.ai/responses/examples)
- [💻 CLI Documentation](https://docs.julep.ai/responses/cli)
- [🔮 Project Roadmap](https://docs.julep.ai/responses/roadmap)

## 📖 About Julep AI

Open Responses is proudly built by [Julep AI](https://julep.ai)—the open-source platform empowering data and ML teams to rapidly create, deploy, and manage stateful AI workflows and intelligent agents at scale.

## 🤝 Contributing

We’d love your contributions! Open Responses is licensed under Apache-2.0.

[Check out our main project as well](https://github.com/julep-ai/julep)

