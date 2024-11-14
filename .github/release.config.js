module.exports = {
    branches: ['main'],
    plugins: [
        ['@semantic-release/commit-analyzer', {
            preset: 'conventionalcommits',
            releaseRules: [
                { type: 'feat', release: 'minor' },
                { type: 'fix', release: 'patch' },
                { scope: 'breaking', release: 'major' },
            ],
        }],
        '@semantic-release/release-notes-generator',
        ['@semantic-release/changelog', {
            changelogFile: 'CHANGELOG.md',
        }],
        ['@semantic-release/exec', {
            prepareCmd: 'mkdir -p release',
        }],
        ['@semantic-release/github', {
            assets: [
                { path: 'release/evm-exporter-linux-amd64' },
                { path: 'release/evm-exporter-darwin-amd64' },
                { path: 'release/evm-exporter-darwin-arm64' },
                { path: 'release/evm-exporter-windows-amd64.exe' }
            ],
        }],
        ['@semantic-release/git', {
            assets: ['CHANGELOG.md'],
            message: 'chore(release): ${nextRelease.version} [skip ci]\n\n${nextRelease.notes}'
        }]
    ]
};
