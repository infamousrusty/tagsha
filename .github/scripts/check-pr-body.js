const github = require('@actions/github');

async function run() {
  const token = process.env.GITHUB_TOKEN;
  const octokit = github.getOctokit(token);
  const context = github.context;

  const pr = context.payload.pull_request;
  const body = (pr.body || '').trim();

  if (body.length >= 20) {
    console.log('PR body is valid.');
    return;
  }

  const files = await octokit.paginate(
    octokit.rest.pulls.listFiles,
    {
      owner: context.repo.owner,
      repo: context.repo.repo,
      pull_number: pr.number,
      per_page: 100
    }
  );

  const names = files.map(function (file) {
    return file.filename;
  });

  const adrFiles = names.filter(function (name) {
    return name.startsWith('adr/') && name.endsWith('.md');
  });

  const adrOnly = names.length > 0 && adrFiles.length === names.length;

  if (!adrOnly) {
    console.error('::error::PR description must be at least 20 characters. Please describe your changes.');
    process.exit(1);
  }

  const lines = [
    'This pull request introduces or modifies the following ADR files:',
    ''
  ];

  adrFiles.forEach(function (name) {
    lines.push('- ' + name);
  });

  lines.push('');
  lines.push('Auto-generated description to satisfy the repository minimum PR description requirement.');

  await octokit.rest.pulls.update({
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: pr.number,
    body: lines.join('
')
  });

  console.log('PR body auto-filled for ADR-only pull request.');
}

run();