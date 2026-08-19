const fs = require('fs');
const path = require('path');

const apiDir = path.join(__dirname, '../lib/api');
const files = fs.readdirSync(apiDir).filter(f => f.endsWith('.ts') && f !== 'client.ts');

for (const file of files) {
  const filePath = path.join(apiDir, file);
  let content = fs.readFileSync(filePath, 'utf8');

  // Remove import "server-only"
  content = content.replace(/import "server-only"\r?\n?/g, '');
  
  // Replace apiFetch import with apiRequest
  content = content.replace(
    /import \{ apiFetch \} from "@\/lib\/api\/client"/g,
    'import { apiRequest } from "@/react-query/client"\nimport { z } from "zod"'
  );

  // Replace get methods
  content = content.replace(
    /const (data|res) = await apiFetch\(([^,]+)\)\r?\n\s*return ([\w.]+)\.parse\(\1\)/g,
    'return apiRequest({ method: "GET", url: $2 }, $3)'
  );

  // Replace delete without return
  content = content.replace(
    /await apiFetch\(([^,]+),\s*\{\s*method:\s*"DELETE"\s*\}\)/g,
    'return apiRequest({ method: "DELETE", url: $1 }, z.any())'
  );

  // Replace post without body and no return
  content = content.replace(
    /await apiFetch\(([^,]+),\s*\{\s*method:\s*"POST"\s*\}\)/g,
    'return apiRequest({ method: "POST", url: $1 }, z.any())'
  );

  // Replace post with body JSON
  content = content.replace(
    /const data = await apiFetch\(([^,]+),\s*\{\s*method:\s*"POST",\s*body:\s*JSON\.stringify\(([^)]+)\),?\s*\}\)\r?\n\s*return ([\w.]+)\.parse\(data\)/g,
    'return apiRequest({ method: "POST", url: $1, data: $2 }, $3)'
  );

  // Replace patch with body JSON
  content = content.replace(
    /const data = await apiFetch\(([^,]+),\s*\{\s*method:\s*"PATCH",\s*body:\s*JSON\.stringify\(([^)]+)\),?\s*\}\)\r?\n\s*return ([\w.]+)\.parse\(data\)/g,
    'return apiRequest({ method: "PATCH", url: $1, data: $2 }, $3)'
  );

  fs.writeFileSync(filePath, content);
}
console.log("Migrated files");
