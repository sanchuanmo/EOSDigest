#include <iostream>
#include <iomanip>
#include <string>
#include "./lib/openssl/include/openssl/sha.h"
#include <vector>
#include "sha256.hpp"
#include "transactionargs.hpp"

using namespace std;


void pack(sha256::encoder& enc, uint8_t in){
  enc.write((char*)&in, sizeof(in));
}

void pack(sha256::encoder& enc, uint32_t in){
  enc.write((char*)&in,sizeof(in));
}

void pack(sha256::encoder& enc, vector<unsigned char>& in){
  enc.write((char*)&in.front(),(uint32_t)in.size());
}

void pack(sha256::encoder& enc, vector<char>& in){
  enc.write(&in.front(),(uint32_t)in.size());
}

void pack227(sha256::encoder& enc, uint32_t in){
  uint64_t val = in;
  do {
    uint8_t b = uint8_t(val) & 0x7f;
    val >>= 7;
    b |= ((val > 0) << 7);
    enc.write((char*)&b,1);
  }while(val);
}

void pack613(sha256::encoder& enc, vector<string>& in){
  pack227(enc,(uint32_t)in.size());

  for( int i = 0 ; i < in.size(); i++){
    pack227(enc,(uint32_t)i);
    vector<unsigned char> signature;
    transSignatrue(in[i],signature);
    pack(enc,signature);
  }
}
void pack306(sha256::encoder& enc, vector<unsigned char>& in){
  pack227(enc,(uint32_t)in.size());

  if (in.size()){
    enc.write((char*)&in.front(),(uint32_t)in.size());
  }
}

void pack(sha256::encoder& enc,const sha256& in){
  enc.write(in.data(),(uint32_t)in.data_size());
}



sha256 Digest(const string status,const uint32_t cpu_usage,const uint32_t net_usage_words,const sha256& digest_packedTrx){
  sha256::encoder enc;
  pack(enc,tranStatus(status));
  pack(enc,cpu_usage);
  pack227(enc,net_usage_words);
  pack(enc,digest_packedTrx);
  return enc.result();
}

sha256 packedDigest(const string compression,vector<unsigned char> &packed_trx, const sha256& digest_prunable){
  sha256::encoder enc;
  pack(enc,tranCompre(compression));
  pack306(enc,packed_trx);
  pack(enc,digest_prunable);
  return enc.result();
}


sha256 prunableDigest(vector<string> &signatures, vector<unsigned char> &context_free_data){
  sha256::encoder enc;
  pack613(enc,signatures);
  pack306(enc,context_free_data);
  return enc.result();
}



string eosServializationTxDigest(const string status,const unsigned int cpu_usage_us,const unsigned int net_usage_words,const string compression,vector<unsigned char>& packed_trx,vector<string>& signatures, vector<unsigned char>& context_free_data){
  sha256 prunable = prunableDigest(signatures, context_free_data);
  sha256 packed = packedDigest(compression,packed_trx,prunable);
  sha256 digest = Digest(status,cpu_usage_us,net_usage_words,packed);
  return string(digest);
}


