k=0
function TestLoop()
    for i=1, 10000, 1 do
        for j=1, 10000, 1 do
            k=k+1
            k=k+k+1
            k=k+k+k+1
            k=k+k+k+k+1
            k=k+k+k+k+k+1
        end
    end
end
print 'haha'
TestLoop();
