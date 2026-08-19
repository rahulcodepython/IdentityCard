const fs = require('fs');
const path = require('path');

const apiDir = path.join(__dirname, '../lib/api');
const files = fs.readdirSync(apiDir).filter(f => f.endsWith('.ts') && f !== 'client.ts');

for (const file of files) {
  const filePath = path.join(apiDir, file);
  let content = fs.readFileSync(filePath, 'utf8');

  // Remove import "server-only"
  content = content.replace(/import "server-only"\r?\n?/g, '');
  
  // Replace import from client
  content = content.replace(
    /import \{ apiFetch \} from "@\/lib\/api\/client"/g,
    'import { apiRequest } from "@/react-query/client"\nimport { z } from "zod"'
  );

  // Replace GET request
  // const data = await apiFetch("/events")
  // return eventsListResponseSchema.parse(data)
  content = content.replace(
    /const (data|res) = await apiFetch\(([^,]+)\)\r?\n\s*return ([\w.]+)\.parse\(\1\)/g,
    'return apiRequest({ method: "GET", url: $2 }, $3)'
  );

  // Replace DELETE without return
  content = content.replace(
    /await apiFetch\(([^,]+),\s*\{\s*method:\s*"DELETE"\s*\}\)/g,
    'return await apiRequest({ method: "DELETE", url: $1 }, z.any())'
  );

  // Replace POST without body and without return
  content = content.replace(
    /await apiFetch\(([^,]+),\s*\{\s*method:\s*"POST"\s*\}\)/g,
    'return await apiRequest({ method: "POST", url: $1 }, z.any())'
  );

  // Replace POST without body WITH return
  content = content.replace(
    /const (data|res) = await apiFetch\(([^,]+),\s*\{\s*method:\s*"POST"\s*\}\)\r?\n\s*return ([\w.]+)\.parse\(\1\)/g,
    'return apiRequest({ method: "POST", url: $2 }, $3)'
  );

  // Replace POST with JSON body
  content = content.replace(
    /const (data|res) = await apiFetch\(([^,]+),\s*\{\s*method:\s*"POST",\s*body:\s*JSON\.stringify\(([^)]+)\),?\s*\}\)\r?\n\s*return ([\w.]+)\.parse\(\1\)/g,
    'return apiRequest({ method: "POST", url: $2, data: $3 }, $4)'
  );

  // Replace POST with FormData body
  content = content.replace(
    /const (data|res) = await apiFetch\(([^,]+),\s*\{\s*method:\s*"POST",\s*body:\s*([^,]+),?\s*\}\)\r?\n\s*return ([\w.]+)\.parse\(\1\)/g,
    'return apiRequest({ method: "POST", url: $2, data: $3 }, $4)'
  );

  // Replace PATCH with JSON body
  content = content.replace(
    /const (data|res) = await apiFetch\(([^,]+),\s*\{\s*method:\s*"PATCH",\s*body:\s*JSON\.stringify\(([^)]+)\),?\s*\}\)\r?\n\s*return ([\w.]+)\.parse\(\1\)/g,
    'return apiRequest({ method: "PATCH", url: $2, data: $3 }, $4)'
  );
  
  // Replace PATCH without return
  content = content.replace(
    /await apiFetch\(([^,]+),\s*\{\s*method:\s*"PATCH",\s*body:\s*JSON\.stringify\(([^)]+)\),?\s*\}\)/g,
    'return await apiRequest({ method: "PATCH", url: $1, data: $2 }, z.any())'
  );

  // Finally, replace POST without return (FormData or generic body)
  content = content.replace(
    /await apiFetch\(([^,]+),\s*\{\s*method:\s*"POST",\s*body:\s*([^,]+),?\s*\}\)/g,
    'return await apiRequest({ method: "POST", url: $1, data: $2 }, z.any())'
  );

  fs.writeFileSync(filePath, content);
}
console.log("Migrated files manually again");
