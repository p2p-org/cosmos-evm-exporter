module.exports = {
    branches: ['main'],
    plugins: [
        '@semantic-release/commit-analyzer',
        '@semantic-release/release-notes-generator',
        '@semantic-release/changelog',
        ['@semantic-release/exec', {
            prepareCmd: 'mkdir -p release',
        }],
        ['@semantic-release/github', {
            assets: [
                { path: 'release/berachain-exporter-linux-amd64', label: 'Linux AMD64 Binary' },
                { path: 'release/berachain-exporter-darwin-amd64', label: 'macOS AMD64 Binary' },
                { path: 'release/berachain-exporter-darwin-arm64', label: 'macOS ARM64 Binary' },
                { path: 'release/berachain-exporter-windows-amd64.exe', label: 'Windows AMD64 Binary' }
            ]
        }],
        ['@semantic-release/git', {
            assets: ['CHANGELOG.md'],
            message: 'chore(release): ${nextRelease.version}\n\n${nextRelease.notes}'
        }]
    ]
};
