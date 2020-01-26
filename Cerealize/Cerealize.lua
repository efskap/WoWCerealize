NumPixels = 64
Interval = 0.2
HeaderBytes = 2

function Cerealize_Send(s)
	for i=1, strlen(s) do
		Cerealize.buf[#Cerealize.buf + 1] = strbyte(s, i)
	end
	Cerealize.buf[#Cerealize.buf + 1] = strbyte('\n')
end

function Cerealize_Tick(self)
	local bufferBytes = Cerealize.buf

	-- remove as many bytes from the buffer as we are able to send
	local maxMsgLen = NumPixels * 3 - HeaderBytes
	local actualMsgLen = min(maxMsgLen, #bufferBytes)
	local bytes = {}
	for i = 1, actualMsgLen do
		bytes[i] = tremove(bufferBytes, 1)
	end
		
	if #bytes > 0 then
		Cerealize.msgNum = Cerealize.msgNum + 1
		if Cerealize.msgNum > 0xF then Cerealize.msgNum = 0 end

		local checksum = 0
		for i = 1, #bytes do
			checksum = checksum + bytes[i]
		end
		checksum = bit.band(checksum, 0xFF)
		table.insert(bytes, 1, checksum)
		table.insert(bytes, 1, Cerealize.msgNum)

		--print('Bytes to send: ', table.concat(bytes, ", ") )
		--print('Buffer size: ', #Cerealize.buf)

		-- now we need to assign each pixel's color to a byte
		for i = 1, NumPixels do
			local curPixelBytes = {} -- 3-length array of current pixel's bytes
			for j=1, 3 do
				local byte = bytes[(i-1)*3 + j]
				if byte == nil then byte = 0 end
				curPixelBytes[j] = byte
			end 
			local t = Cerealize.pixels[i]
			t:SetColorTexture(curPixelBytes[1]/255,curPixelBytes[2]/255,curPixelBytes[3]/255)
		end
	end
end



function Cerealize_OnLoad(self)
	local scale = string.match( GetCVar( "gxWindowedResolution" ), "%d+x(%d+)" );
	local uiScale = UIParent:GetScale( );
	Cerealize:SetScale(768/scale/uiScale);

	Cerealize.pixels = {}
	Cerealize.msgNum = 0
	Cerealize.buf = {}
	Cerealize_Send('INIT!!!')
	for i = 1, NumPixels do
		local t = Cerealize:CreateTexture(nil,"DIALOG",nil,1)
		t:SetPoint("BOTTOMRIGHT", -NumPixels + i, 0)
		t:SetSize(1,1)
		Cerealize.pixels[i] = t
	end
	-- align with system clock to the second
	local startTime = time()
	bootstrapTicker = C_Timer.NewTicker(0.01, function(self)
		print("wait")
		if time() > startTime then
			self:Cancel()
			C_Timer.NewTicker(Interval, Cerealize_Tick)
		end

	end)
	Cerealize:RegisterAllEvents();
	local function eventHandler(self, ...)
		Cerealize_Send(tableToJson({...}))
	end
	Cerealize:SetScript("OnEvent", eventHandler);
end

function tableToJson(tableWithData)
	local result = {}

	for key, value in ipairs(tableWithData) do
		-- prepare json key-value pairs and save them in separate table
		table.insert(result, string.format("\"%s\":%s", key, tostring(value)))
	end

	-- get simple json string
	result = "{" .. table.concat(result, ",") .. "}"
	return result
end
