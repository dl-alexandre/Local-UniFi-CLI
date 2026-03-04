#!/usr/bin/env python3
"""Add context creation to CLI command Run methods"""

import re

# Read the file
with open('internal/pkg/cli/cli.go', 'r') as f:
    content = f.read()

# Pattern to find Run methods that call g.initClient() followed by g.resolveSiteID
# We need to add context creation after g.initClient() and before g.resolveSiteID

# Pattern 1: Find "if err := g.initClient(); err != nil { return err }" followed by "siteID, err := g.resolveSiteID"
pattern1 = r'(func \(c \*\w+Cmd\) Run\(g \*Globals\) error \{\s*if err := g\.initClient\(\); err != nil \{\s*return err\s*\}\s*)(siteID, err := g\.resolveSiteID\(ctx, c\.Site\))'

replacement1 = r'''\1// Create context with timeout
\tctx, cancel := g.createContext()
\tdefer cancel()

\t\2'''

content = re.sub(pattern1, replacement1, content, flags=re.MULTILINE)

# Pattern 2: Also need to handle PingCmd which directly calls g.appClient.ListSites without resolveSiteID
pattern2 = r'(func \(c \*PingCmd\) Run\(g \*Globals\) error \{\s*if err := g\.initClient\(\); err != nil \{\s*return err\s*\}\s*)(// Try to list sites)'
replacement2 = r'''\1// Create context with timeout
\tctx, cancel := g.createContext()
\tdefer cancel()

\t\2'''

content = re.sub(pattern2, replacement2, content, flags=re.MULTILINE)

# Pattern 3: Also need to handle ListSitesCmd which directly calls g.appClient.ListSites without resolveSiteID  
pattern3 = r'(func \(c \*ListSitesCmd\) Run\(g \*Globals\) error \{\s*if err := g\.initClient\(\); err != nil \{\s*return err\s*\}\s*)(resp, err := g\.appClient\.ListSites\(\))'
replacement3 = r'''\1// Create context with timeout
\tctx, cancel := g.createContext()
\tdefer cancel()

\t\2'''

content = re.sub(pattern3, replacement3, content, flags=re.MULTILINE)

# Pattern 4: Fix g.appClient.ListSites() to g.appClient.ListSites(ctx)
content = re.sub(r'g\.appClient\.ListSites\(\)', r'g.appClient.ListSites(ctx)', content)

# Pattern 5: Fix g.appClient.ListDevices(siteID) to g.appClient.ListDevices(ctx, siteID)  
content = re.sub(r'g\.appClient\.ListDevices\(([^)]+)\)', r'g.appClient.ListDevices(ctx, \1)', content)

# Pattern 6: Fix g.appClient.ListClients(siteID) to g.appClient.ListClients(ctx, siteID)
content = re.sub(r'g\.appClient\.ListClients\(([^)]+)\)', r'g.appClient.ListClients(ctx, \1)', content)

# Pattern 7: Fix g.appClient.GetSiteHealth(siteID) to g.appClient.GetSiteHealth(ctx, siteID)
content = re.sub(r'g\.appClient\.GetSiteHealth\(([^)]+)\)', r'g.appClient.GetSiteHealth(ctx, \1)', content)

# Pattern 8: Fix g.appClient.AdoptDevice(siteID, c.MAC) to g.appClient.AdoptDevice(ctx, siteID, c.MAC)
content = re.sub(r'g\.appClient\.AdoptDevice\(([^,]+),\s*([^)]+)\)', r'g.appClient.AdoptDevice(ctx, \1, \2)', content)

# Write the file back
with open('internal/pkg/cli/cli.go', 'w') as f:
    f.write(content)

print("Updated CLI file with context support")
