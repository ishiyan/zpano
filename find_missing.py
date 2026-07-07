import re

# Get all public methods from ratios.py
with open('py/performance/ratios.py', 'r') as f:
    content = f.read()

methods = re.findall(r'^\s+def (\w+)', content, re.MULTILINE)
properties = re.findall(r'^\s+@property\n^\s+def (\w+)', content, re.MULTILINE)
all_methods = set(m for m in methods + properties if not m.startswith('_') and m not in ['__init__', 'reset', 'add_return', 'finalize_calculation'])

# Get all methods mentioned in implementation-status.md with braverock and testdata
with open('external/implementation-status.md', 'r') as f:
    md_content = f.read()

# Find rows with both braverock and ratios but no testdata
table_rows = re.findall(r'\|\s*\[(.*?)\]\(.*?\)\s*\|\s*(.*?)\s*\|\s*(.*?)\s*\|\s*([^|]+)\s*\|', md_content)

print('Methods with Braverock but NO testdata:')
for bacon3, braverock, ratios, testdata in table_rows:
    braverock = braverock.strip()
    ratios = ratios.strip()
    testdata = testdata.strip()
    
    if braverock and braverock != '-' and ratios and ratios != '-' and testdata not in ['✅', '✅ (rf 0-0.3 by 0.05)', '✅ (StdDev, VaR, ES, SemiSD; rf 0-0.3 by 0.05)', '✅ (geometric=True/False)', '✅ (l=1,2,3,4; MAR 0-0.3 by 0.05)', '✅ (MAR 0-0.3 by 0.05)', '✅ (excess 0-0.1 by 0.02)', '✅ (L 0-0.1 by 0.02)', '✅ (full/subset; MAR 0-0.1 by 0.02)']:
        if '❌' in testdata or testdata == '-' or testdata == '' or '(commented' in testdata:
            print(f'  {ratios} | {braverock} | {testdata}')