import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import sitemap from '@astrojs/sitemap';

export default defineConfig({
  site: 'https://proxysql.github.io',
  base: '/dbdeployer',
  integrations: [
    starlight({
      title: 'dbdeployer',
      description: 'Deploy MySQL & PostgreSQL sandboxes in seconds',
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/ProxySQL/dbdeployer' },
      ],
      sidebar: [
        {
          label: 'Getting Started',
          items: [
            { label: 'Installation', slug: 'getting-started/installation' },
            { label: 'Quick Start: MySQL Single', slug: 'getting-started/quickstart-mysql-single' },
            { label: 'Quick Start: MySQL Replication', slug: 'getting-started/quickstart-mysql-replication' },
            { label: 'Quick Start: PostgreSQL', slug: 'getting-started/quickstart-postgresql' },
            { label: 'Quick Start: ProxySQL Integration', slug: 'getting-started/quickstart-proxysql' },
          ],
        },
        {
          label: 'Core Concepts',
          items: [
            { label: 'Sandboxes', slug: 'concepts/sandboxes' },
            { label: 'Versions & Flavors', slug: 'concepts/flavors' },
            { label: 'Ports & Networking', slug: 'concepts/ports' },
            { label: 'Environment Variables', slug: 'concepts/environment-variables' },
          ],
        },
        {
          label: 'Deploying',
          items: [
            { label: 'Single Sandbox', slug: 'deploying/single' },
            { label: 'Multiple Sandboxes', slug: 'deploying/multiple' },
            { label: 'Replication', slug: 'deploying/replication' },
            { label: 'Group Replication', slug: 'deploying/group-replication' },
            { label: 'Fan-In & All-Masters', slug: 'deploying/fan-in-all-masters' },
            { label: 'NDB Cluster', slug: 'deploying/ndb-cluster' },
            { label: 'InnoDB Cluster', slug: 'deploying/innodb-cluster' },
          ],
        },
        {
          label: 'Providers',
          items: [
            { label: 'MySQL', slug: 'providers/mysql' },
            { label: 'PostgreSQL', slug: 'providers/postgresql' },
            { label: 'ProxySQL', slug: 'providers/proxysql' },
            { label: 'Percona XtraDB Cluster', slug: 'providers/pxc' },
          ],
        },
        {
          label: 'Managing Sandboxes',
          items: [
            { label: 'Starting & Stopping', slug: 'managing/starting-stopping' },
            { label: 'Using Sandboxes', slug: 'managing/using' },
            { label: 'Customization', slug: 'managing/customization' },
            { label: 'Database Users', slug: 'managing/users' },
            { label: 'Logs', slug: 'managing/logs' },
            { label: 'Deletion & Cleanup', slug: 'managing/deletion' },
            { label: 'Admin Web UI', slug: 'managing/admin-ui' },
          ],
        },
        {
          label: 'Advanced',
          items: [
            { label: 'Concurrent Deployment', slug: 'advanced/concurrent' },
            { label: 'Importing Databases', slug: 'advanced/importing' },
            { label: 'Inter-Sandbox Replication', slug: 'advanced/inter-sandbox-replication' },
            { label: 'Cloning', slug: 'advanced/cloning' },
            { label: 'Using as a Go Library', slug: 'advanced/go-library' },
            { label: 'Compiling from Source', slug: 'advanced/compiling' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'CLI Commands', slug: 'reference/cli-commands' },
            { label: 'Configuration', slug: 'reference/configuration' },
            { label: 'Topology Reference', slug: 'reference/topology-reference' },
            { label: 'API Changelog', slug: 'reference/api-changelog' },
          ],
        },
      ],
    }),
    sitemap(),
  ],
});
