#!/bin/bash

echo "=========================================="
echo "Quick CPU Test - 20 seconds total"
echo "=========================================="
echo ""
echo "Opening htop in 2 seconds..."
echo "Look for 'go run' process"
sleep 2

# Open htop in background if available
if command -v htop &> /dev/null; then
    htop &
    HTOP_PID=$!
elif command -v top &> /dev/null; then
    top &
    HTOP_PID=$!
fi

sleep 2

echo ""
echo "Running demo..."
go run simple_cpu_demo.go

# Close htop
if [ ! -z "$HTOP_PID" ]; then
    kill $HTOP_PID 2>/dev/null
fi

echo ""
echo "Test complete! Did you see the difference?"
echo "- Busy-wait: 100% CPU"
echo "- Blocking: ~0% CPU"
