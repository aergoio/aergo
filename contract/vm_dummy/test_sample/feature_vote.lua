function constructor()
	system.setItem("owner", system.getSender())
end

function addCandidate(name)
	if system.getSender() ~= system.getItem("owner") then
		return
	end

	if (system.getItem(name) ~= nil) then
		return
	end
	
	local numCandidates;
	if (system.getItem("numCandidates") == nil) then
		numCandidates = 0;
	else
		numCandidates = tonumber(system.getItem("numCandidates"))
	end

	system.setItem("candidate_list_" .. numCandidates, name)

	numCandidates = numCandidates + 1;
	system.setItem("numCandidates", tostring(numCandidates));
	system.setItem(name, tostring(0));
end

function getCandidates()
	local numCandidates;
	if (system.getItem("numCandidates") == nil) then
		return {};
	else
		numCandidates = tonumber(system.getItem("numCandidates"))
	end

	local candidates = {};
	local i = 0;

	while true do
		if (numCandidates == i) then
			break;
		end
		local candidate = system.getItem("candidate_list_" .. i)
		local count = system.getItem(candidate)
		if count == nil then
			count = 0
		end
		table.insert(candidates, {id = i, name = candidate, count = count});
		i = i + 1;
	end
	return candidates;
end

function registerVoter(address)
	if system.getSender() ~= system.getItem("owner") then
		return
	end
	
	system.setItem("voter_" .. address, "0");
end

function vote(candidateID)
	local totalVoted
	local voter = system.getItem("voter_" .. system.getSender())
	if voter == nil then
		return
	end
	totalVoted = tonumber(system.getItem("voter_" .. system.getSender()))
	if totalVoted > 3 then
		return
	end
	if system.getItem(candidateID) == nil then
		return
	end
	local currentVotes;
	if (system.getItem(candidateID) == nil) then
		currentVotes = 0;
	else
		currentVotes = tonumber(system.getItem(candidateID))
	end
	currentVotes = currentVotes + 1

	system.setItem(candidateID, tostring(currentVotes))
	totalVoted = totalVoted + 1
	system.setItem("voter_" .. system.getSender(), tostring(totalVoted));
end

abi.register(addCandidate, getCandidates, registerVoter, vote)