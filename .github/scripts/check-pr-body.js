const https = require('https');
const fs = require('fs');

const NEWLINE = String.fromCharCode(10);

function sanitiseString(value) {
  if (typeof value !== 'string') { return ''; }
  return value.replace(/[^ws-/.]/g, '').trim();
}

function sanitiseFilename(value) {
  if (typeof value !== 'string') { return null; }
  const clean = value.replace(/[^a-zA-Z0-9_-./]/g, '').trim();
  if (clean.length === 0) { return null; }
  return clean;
}

function apiRequest(method, path, body) {
  return new Promise(function (resolve, reject) {
    const data = body ? JSON.stringify(body) : null;
    const options = {
      hostname: 'api.github.com',
      path: path,
      method: method,
      headers: {
        'Authorization': 'Bearer ' + process.env.GITHUB_TOKEN,
        'Accept': 'application/vnd.github+json',
        'X-GitHub-Api-Version': '2022-11-28',
        'User-Agent': 'tagsha-pr-validation',
        'Content-Type': 'application/json'
      }
    };

    if (data) {
      options.headers['Content-Length'] = Buffer.byteLength(data);
    }

    const req = https.request(options, function (res) {
      let raw = '';
      res.on('data', function (chunk) { raw += chunk; });
      res.on('end', function () {
        try { resolve(JSON.parse(raw)); }
        catch (e) { resolve({}); }
      });
    });

    req.on('error', reject);
    if (data) { req.write(data); }
    req.end();
  });
}

async function getFiles(owner, repo, prNumber) {
  const results = [];
  let page = 1;

  while (true) {
    const path = '/repos/' + owner + '/' + repo + '/pulls/' + prNumber + '/files?per_page=100&page=' + page;
    const batch = await apiRequest('GET', path, null);
    if (!Array.isArray(batch) || batch.length === 0) { break; }
    results.push.apply(results, batch);
    if (batch.length < 100) { break; }
    page++;
  }

  return results;
}

async function updatePrBody(owner, repo, prNumber, newBody) {
  const path = '/repos/' + owner + '/' + repo + '/pulls/' + prNumber;
  await apiRequest('PATCH', path, { body: newBody });
}

async function run() {
  const eventRaw = fs.readFileSync(process.env.GITHUB_EVENT_PATH, 'utf8');
  const payload = JSON.parse(eventRaw);
  const pr = payload.pull_request;

  // Sanitise all values sourced from the event file before use
  const prBody = (typeof pr.body === 'string' ? pr.body : '').trim();
  const prNumber = parseInt(pr.number, 10);

  const repoParts = process.env.GITHUB_REPOSITORY.split('/');
  const owner = sanitiseString(repoParts[0]);
  const repo = sanitiseString(repoParts[1]);

  if (!owner || !repo || isNaN(prNumber)) {
    console.log('::error::Could not determine repository owner, name, or PR number.');
    process.exit(1);
  }

  if (prBody.length >= 20) {
    console.log('PR body is valid.');
    process.exit(0);
  }

  const files = await getFiles(owner, repo, prNumber);

  // Sanitise all filenames from the API response before use in request body
  const names = files
    .map(function (file) { return sanitiseFilename(file.filename); })
    .filter(function (name) { return name !== null; });

  const adrFiles = names.filter(function (name) {
    return name.startsWith('adr/') && name.endsWith('.md');
  });

  const adrOnly = names.length > 0 && adrFiles.length === names.length;

  if (!adrOnly) {
    console.log('::error::PR description must be at least 20 characters. Please describe your changes.');
    process.exit(1);
  }

  const lines = [
    'This pull request introduces or modifies the following ADR files:',
    ''
  ];

  adrFiles.forEach(function (name) { lines.push('- ' + name); });
  lines.push('');
  lines.push('Auto-generated description to satisfy the repository minimum PR description requirement.');

  await updatePrBody(owner, repo, prNumber, lines.join(NEWLINE));
  console.log('PR body auto-filled for ADR-only pull request.');
  process.exit(0);
}

run();