---
title: "API Reference"
description: "REST API endpoints and usage"
weight: 320
icon: "code"
toc: true
---

# API Documentation

## Overview

The Nom Database API provides comprehensive endpoints for managing restaurants, ratings, categories, and food types, with Google Maps integration for location-based features.

## Interactive API Documentation

### Swagger UI

Access the interactive Swagger UI documentation at:

**Development**: [http://localhost:8080/api/docs](http://localhost:8080/api/docs)

The Swagger UI provides:
- Interactive endpoint testing
- Request/response examples
- Schema definitions
- Try-it-out functionality for all endpoints

### OpenAPI Specification

The raw OpenAPI specification is available at:

**YAML Format**: [http://localhost:8080/api/swagger.yaml](http://localhost:8080/api/swagger.yaml)
**JSON Format**: `http://localhost:8080/api/swagger.json` (via docs endpoint)

## Quick Start

### Making API Requests

All API requests should be made to:
```
http://localhost:8080/api
```

### Response Format

All responses are in JSON format with appropriate HTTP status codes.

Success Response Example:
```json
{
  "id": 1,
  "name": "Pizza Palace",
  "created_at": "2025-12-30T12:00:00Z"
}
```

Error Response Example:
```json
{
  "error": "Restaurant not found"
}
```

## Available Endpoints

### Restaurants

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/restaurants` | List all restaurants with optional filters |
| `GET` | `/restaurants/{id}` | Get restaurant details by ID |
| `POST` | `/restaurants` | Create a new restaurant |
| `PUT` | `/restaurants/{id}` | Update a restaurant |
| `DELETE` | `/restaurants/{id}` | Delete a restaurant |
| `GET` | `/restaurants/paginated` | Get paginated list of restaurants |
| `GET` | `/search` | Global search across restaurants |

### Ratings

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/restaurants/{restaurantId}/ratings` | Get all ratings for a restaurant |
| `POST` | `/ratings` | Create a new rating |
| `DELETE` | `/ratings/{id}` | Delete a rating |

### Categories

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/categories` | List all categories |
| `GET` | `/categories/{id}` | Get category by ID |
| `POST` | `/categories` | Create a new category |
| `PUT` | `/categories/{id}` | Update a category |
| `DELETE` | `/categories/{id}` | Delete a category |

### Food Types

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/food-types` | List all food types |
| `GET` | `/food-types/{id}` | Get food type by ID |
| `POST` | `/food-types` | Create a new food type |
| `PUT` | `/food-types/{id}` | Update a food type |
| `DELETE` | `/food-types/{id}` | Delete a food type |

### Suggestions

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/suggestions` | List all restaurant suggestions |
| `GET` | `/suggestions/{id}` | Get suggestion by ID |
| `POST` | `/suggestions` | Create a new suggestion |
| `PATCH` | `/suggestions/{id}/status` | Update suggestion status |
| `POST` | `/suggestions/{id}/convert` | Convert suggestion to restaurant |
| `DELETE` | `/suggestions/{id}` | Delete a suggestion |

### Google Maps Integration

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/places/search` | Search for places using Google Maps |
| `GET` | `/places/{placeId}` | Get detailed place information |
| `GET` | `/geocode/cities` | Geocode cities |

### Menu Photos

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/restaurants/{restaurantId}/photos` | Get all photos for a restaurant |
| `POST` | `/restaurants/{restaurantId}/photos` | Upload a menu photo |
| `PATCH` | `/photos/{id}` | Update photo caption |
| `DELETE` | `/photos/{id}` | Delete a photo |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Server health check |

## Endpoint Examples

### List Restaurants with Filters

```bash
# Get all restaurants
curl http://localhost:8080/api/restaurants

# Filter by category
curl http://localhost:8080/api/restaurants?category_id=1

# Filter by food types (comma-separated)
curl http://localhost:8080/api/restaurants?food_type_ids=1,2,3

# Filter by location (within radius)
curl "http://localhost:8080/api/restaurants?lat=40.7128&lng=-74.0060&radius=5"
```

### Get Restaurant Details

```bash
curl http://localhost:8080/api/restaurants/1
```

Response:
```json
{
  "id": 1,
  "name": "Pizza Palace",
  "description": "Authentic Italian pizza",
  "address": "123 Main St, New York, NY",
  "phone": "555-0100",
  "website": "https://pizzapalace.com",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "google_place_id": "ChIJ...",
  "category_id": 1,
  "category": {
    "id": 1,
    "name": "Italian"
  },
  "food_types": [
    {
      "id": 1,
      "name": "Pizza"
    }
  ],
  "avg_rating": {
    "food": 4.5,
    "service": 4.2,
    "ambiance": 4.3,
    "overall": 4.33,
    "count": 10
  },
  "created_at": "2025-12-30T12:00:00Z",
  "updated_at": "2025-12-30T12:00:00Z"
}
```

### Create a Restaurant

```bash
curl -X POST http://localhost:8080/api/restaurants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sushi Bar",
    "description": "Fresh sushi and sashimi",
    "address": "456 Oak Ave, New York, NY",
    "phone": "555-0200",
    "website": "https://sushibar.com",
    "latitude": 40.7580,
    "longitude": -73.9855,
    "category_id": 2,
    "food_type_ids": [4, 5]
  }'
```

### Create a Rating

```bash
curl -X POST http://localhost:8080/api/ratings \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_id": 1,
    "food_rating": 5,
    "service_rating": 4,
    "ambiance_rating": 4,
    "comment": "Great pizza, friendly staff!"
  }'
```

### Search Places (Google Maps)

```bash
curl "http://localhost:8080/api/places/search?q=pizza+new+york"
```

### Paginated Restaurants

```bash
# First page
curl http://localhost:8080/api/restaurants/paginated?limit=20

# Next page using cursor
curl "http://localhost:8080/api/restaurants/paginated?limit=20&cursor=eyJpZCI6MjB9"
```

Response:
```json
{
  "data": [...],
  "next_cursor": "eyJpZCI6NDB9",
  "has_more": true,
  "total": 100
}
```

## Data Models

### Restaurant

```json
{
  "id": integer,
  "name": string,
  "description": string,
  "address": string,
  "phone": string,
  "website": string,
  "latitude": number,
  "longitude": number,
  "google_place_id": string,
  "category_id": integer,
  "category": Category,
  "food_types": [FoodType],
  "avg_rating": AvgRating,
  "distance": number,
  "is_suggestion": boolean,
  "suggestion_id": integer,
  "status": string,
  "created_at": string,
  "updated_at": string
}
```

### Rating

```json
{
  "id": integer,
  "restaurant_id": integer,
  "food_rating": integer (1-5),
  "service_rating": integer (1-5),
  "ambiance_rating": integer (1-5),
  "comment": string,
  "created_at": string
}
```

### Category

```json
{
  "id": integer,
  "name": string,
  "created_at": string,
  "updated_at": string
}
```

### FoodType

```json
{
  "id": integer,
  "name": string,
  "created_at": string,
  "updated_at": string
}
```

### AvgRating

```json
{
  "food": number,
  "service": number,
  "ambiance": number,
  "overall": number,
  "count": integer
}
```

## HTTP Status Codes

| Code | Description |
|------|-------------|
| `200` | OK - Request successful |
| `201` | Created - Resource created successfully |
| `400` | Bad Request - Invalid request data |
| `404` | Not Found - Resource not found |
| `409` | Conflict - Resource already exists |
| `500` | Internal Server Error - Server error |

## Rate Limiting

The API implements rate limiting to prevent abuse:
- **Limit**: 100 requests per minute per IP address
- **Burst**: 20 requests

When rate limit is exceeded, the API returns:
```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```

## CORS

The API supports Cross-Origin Resource Sharing (CORS) for the following origins:
- `http://localhost:3000` (Frontend dev server)
- `http://localhost:5173` (Vite dev server)

Production origins can be configured via the `ALLOWED_ORIGINS` environment variable.

## Adding Documentation for New Endpoints

When adding new API endpoints, follow these steps to update the Swagger documentation:

### 1. Add Swagger Annotations

Add godoc comments with Swagger annotations above your handler function:

```go
// GetMyEndpoint godoc
// @Summary Brief description
// @Description Detailed description
// @Tags TagName
// @Accept json
// @Produce json
// @Param id path int true "ID parameter description"
// @Param filter query string false "Query parameter description"
// @Success 200 {object} models.MyModel "Success response"
// @Failure 400 {object} map[string]string "Error description"
// @Failure 500 {object} map[string]string "Error description"
// @Router /my-endpoint/{id} [get]
func GetMyEndpoint(w http.ResponseWriter, r *http.Request) {
    // Handler implementation
}
```

### 2. Regenerate Documentation

Run swag init to regenerate the Swagger files:

```bash
cd backend
~/go/bin/swag init -g cmd/server/main.go -o docs
```

### 3. Rebuild and Test

```bash
# Rebuild the backend
docker compose up -d --build backend

# Test the documentation
open http://localhost:8080/api/docs
```

## Swagger Annotation Reference

### Common Tags

- `@Summary`: Short one-line description
- `@Description`: Detailed multi-line description
- `@Tags`: Group endpoints by category
- `@Accept`: Request content type (json, xml, etc.)
- `@Produce`: Response content type
- `@Param`: Parameter definition
  - Format: `name location type required "description"`
  - Locations: path, query, header, body, formData
- `@Success`: Success response
  - Format: `statusCode {type} model "description"`
- `@Failure`: Error response
- `@Router`: Route definition
  - Format: `path [method]`

### Parameter Types

- `string` - String value
- `integer` or `int` - Integer number
- `number` - Floating point number
- `boolean` or `bool` - Boolean value
- `object` - Object type (requires model reference)
- `array` - Array type

### Model References

Use `{object}` or `{array}` with model names:
```go
// @Success 200 {object} models.Restaurant
// @Success 200 {array} models.Category
// @Param body body models.CreateRequest true "Request body"
```

## Testing the API

### Using Swagger UI

1. Navigate to http://localhost:8080/api/docs
2. Browse available endpoints
3. Click "Try it out" on any endpoint
4. Fill in parameters
5. Click "Execute"
6. View the response

### Using curl

See the examples above for curl command examples.

### Using Postman

Import the OpenAPI specification:
1. Open Postman
2. Click Import
3. Paste the URL: `http://localhost:8080/api/swagger.yaml`
4. Postman will create a collection with all endpoints

## Security

The API includes several security features:
- **Rate Limiting**: 100 requests/min per IP
- **Input Sanitization**: XSS and SQL injection prevention
- **Request Size Limits**: 10MB maximum request size
- **Security Headers**: XSS protection, clickjacking prevention
- **Content-Type Validation**: Enforced for POST/PUT requests
- **CORS Restrictions**: Limited to configured origins

## Support

For API issues or questions:
- GitHub Issues: https://github.com/obermarclp/the-nom-database/issues
