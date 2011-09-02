package main
import (
	"net"
	"bufio"
	"container/list"
	"io"
	"json"
	"log"
	"os"
)

type SubscribeRequest struct {
	dataChan chan JSONData
	id int
}

type DataStream struct {
	name string
	connectString string
	
	// We support keeping a cache of recent data items for later inspection. Obviously we want to cap the size on this.
	dataCacheKey string
	dataCacheIndexes list.List
	dataCache map [string] *JSONData

	rawStream io.ReadWriteCloser  // Raw io stream of data
	ioStream *bufio.Reader  // Our buffered view of our data stream
	
	subscribeChan chan *SubscribeRequest
	unsubscribeChan chan *SubscribeRequest
	
	allChannels [](chan JSONData)
}

func NewDataStream(name string, connectString string) (stream *DataStream) {
	stream = new(DataStream)
	stream.name = name
	stream.connectString = connectString

	// TODO: Make this dynamic
	stream.dataCacheKey = "unique_request_id"
	stream.subscribeChan = make(chan *SubscribeRequest)
	stream.unsubscribeChan = make(chan *SubscribeRequest)
	stream.allChannels = make([](chan JSONData), 0, 64)

	stream.dataCache = make(map [string] *JSONData, 64)
	
	stream.dataCacheIndexes.Init()

	go stream.acceptChannels()
	return
}

func (stream *DataStream) acceptChannels() {
	for {
		
		select {
			case channelRequest := <-stream.subscribeChan:
				stream.subscribe(channelRequest)
			case channelRequest := <-stream.unsubscribeChan:
				stream.unsubscribe(channelRequest)
		}
	}
}

func (stream *DataStream) subscribe(request *SubscribeRequest) {
	request.id = -1
	for ndx, value := range(stream.allChannels) {
		if value == nil {
			stream.allChannels[ndx] = request.dataChan
			request.id = ndx
			break
		}
	}
	if request.id < 0 {
		stream.allChannels = append(stream.allChannels, request.dataChan)
		request.id = (len(stream.allChannels) - 1)
	}
	log.Printf("Adding new channel %d to data stream", request.id, stream.name)

	// If we are not yet streaming data, we should be
	if stream.ioStream == nil {
		stream.createIOStream()
		go stream.streamData()
	}
}

func (stream *DataStream) unsubscribe(request *SubscribeRequest) {
	log.Println("Dropping channel", request.id)
	stream.allChannels[request.id] = nil
}

func (stream *DataStream) cacheData(data *JSONData) {
	dataKey, ok := GetDeep(stream.dataCacheKey, *data)
	if !ok {
		log.Printf("Failed to find key %s", stream.dataCacheKey)
		return
	}
	
	dataKeyStr, ok := dataKey.(string)
	if !ok {
		// No key for us
		return
	}
	stream.dataCache[dataKeyStr] = data
	stream.dataCacheIndexes.PushBack(data)

	if stream.dataCacheIndexes.Len() >= 64 {
		stream.dataCacheIndexes.Remove(stream.dataCacheIndexes.Front())
	}

	log.Printf("Data cache now contains %d elements", stream.dataCacheIndexes.Len())
}

func (stream *DataStream) LookupData(key string) *JSONData {
	data, ok := stream.dataCache[key]; 
	if !ok {
		return nil
	}

	return data
}

func (stream *DataStream) streamData() {
	for {
		line, isPrefix, err := stream.ioStream.ReadLine()
		if err != nil {
			if err == os.EOF {
				break
			}
			log.Printf("Failed on line stream", err)
			break
		}
		if isPrefix {
			log.Printf("PREFIX!! Skipping line.")
			continue
		}

		// We have fairly reliable looking chunk of data, try to decode it
		var data JSONData
		err = json.Unmarshal(line, &data)
		if err != nil {
			log.Printf("Failure to decode: %s", err)
			log.Println(string(line))
			log.Println()
			continue
		}

		// Add to our cache
		stream.cacheData(&data)

		// Now deliver this fine chunk of ranger data to each of our listeners
		sent := false
		for ndx, dataChannel := range stream.allChannels {
			if dataChannel != nil {
				// We don't want to be blocking waiting on the channel, if it can't keep up we'll drop the data.
				select {
					case dataChannel <- data:
					default:
						log.Println("Dropping data to channel", ndx)
				}
				sent = true
			}
		}
		/* There are no dataChannel's left open, we can close the stream */
		if !sent {
			log.Printf("Closing data stream for %s", stream.name)
			stream.rawStream.Close()
			stream.rawStream = nil;
			stream.ioStream = nil;
			break
		}
	}
	log.Printf("All done with data stream %s", stream.name)
}


func (stream *DataStream) createIOStream() {
	conn, err := net.Dial("tcp4", stream.connectString)
	if err != nil {
		log.Fatal("Failed to open", err)
	}

	_, err = conn.Write([]uint8(stream.name + "\n"))
	if err != nil {
		log.Fatal("Failed to send cmd", err)
	}

	stream.rawStream = conn
	stream.ioStream, err = bufio.NewReaderSize(conn, 1024*32)
	if err != nil {
		log.Fatal("Failed to create reader", err)
	}
}
