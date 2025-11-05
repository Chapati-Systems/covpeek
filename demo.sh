#!/bin/bash

echo "========================================"
echo "covpeek - Coverage Report Parser Demo"
echo "========================================"
echo ""

echo "1. Parsing Rust Coverage (LCOV format)"
echo "--------------------------------------"
./covpeek parse --file testdata/sample.lcov
echo ""

echo "2. Parsing Go Coverage (.out format)"
echo "--------------------------------------"
./covpeek parse --file testdata/sample.out
echo ""

echo "3. Parsing TypeScript Coverage (LCOV format)"
echo "--------------------------------------"
./covpeek parse --file testdata/typescript.info
echo ""

echo "4. Handling Malformed Input (with warnings)"
echo "--------------------------------------"
./covpeek parse --file testdata/malformed.lcov
echo ""

echo "5. JSON Output Example"
echo "--------------------------------------"
echo "Output (first 20 lines):"
./covpeek parse --file testdata/typescript.info --format json | head -20
echo ""

echo "========================================"
echo "Demo Complete!"
echo "========================================"
