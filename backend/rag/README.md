# eg：

<img width="1554" height="756" alt="截屏2026-01-26 11 28 22" src="https://github.com/user-attachments/assets/f8af17d6-4dfe-4604-bf7b-5c9a9ae1cded" />


# RAG System for CoinWave

This module implements a Retrieval-Augmented Generation (RAG) system using:
- **Go + Gin**: Backend framework
- **Milvus**: Vector Database (for storing article chunks)
- **Volcengine (Doubao)**: Embedding and LLM provider
- **Eino**: Task orchestration for ingestion pipeline

## Features
1. **Private Knowledge Base**: Strict `user_id` isolation ensures users only search their own articles.
2. **Async Ingestion**: Articles are processed asynchronously using an Eino Graph (Fetch -> Chunk -> Embed -> Store).
3. **Hybrid Storage**: Metadata in MySQL/SQLite, Vectors in Milvus.

## Configuration
Configuration is located in `backend/rag/config.go`.
Key settings:
- `VolcAuthToken`: Your Volcengine API Key.
- `MilvusAddress`: Address of your Milvus instance (default `localhost:19530`).

## Dependencies
Ensure you have a running Milvus instance.
You can run Milvus using Docker:
```bash
wget https://github.com/milvus-io/milvus/releases/download/v2.4.0/milvus-standalone-docker-compose.yml -O docker-compose.yml
docker-compose up -d
```

## API Endpoints
- **POST /api/articles**: Create an article (triggers async vectorization).
- **POST /api/rag/query**: Ask a question based on your knowledge base.
  - Header: `Authorization: Bearer <token>`
  - Body: `{"question": "Your question here"}`
- **POST /api/articles/:id/reindex**: Manually trigger re-indexing for an article.
