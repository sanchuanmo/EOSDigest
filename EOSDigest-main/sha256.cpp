#include "sha256.hpp"
#include "utils.hpp"


sha256::sha256() {memset(_hash,0,sizeof(_hash));}

sha256::sha256(const string& hex_str) {
    auto bytes_written = hexToBytes(hex_str, (unsigned char*)_hash);
    if (bytes_written < sizeof(_hash)){
        memset((char*)_hash + bytes_written, 0, (sizeof(_hash) - bytes_written));
    }
}

string sha256::str() const {
    return bytestoHex((char*)_hash,sizeof(_hash));
}

sha256::operator string()const {return str();}

const char* sha256::data()const {return (const char*)&_hash[0];}
char* sha256::data() {return (char*)&_hash[0];}

sha256::encoder::~encoder() {}
sha256::encoder::encoder() {
    reset();
}

sha256 sha256::hash(const char* d, uint32_t dlen) {
    encoder e;
    e.write(d,dlen);
    return e.result();
}

sha256 sha256::hash(const unsigned char* d, uint32_t dlen){
    encoder e;
    e.write((char*)d,dlen);
    return e.result();
}

sha256 sha256::hash(const string& s){
    return hash(s.c_str(), s.size());
}

sha256 sha256::hash(const sha256& s){
    return hash(s.data(), sizeof(s._hash));
}

sha256 sha256::hash(const std::pair<sha256,sha256>&  value){
    encoder e;
    e.write(value.first.data(),sizeof(value.first._hash));
    e.write(value.second.data(),sizeof(value.second._hash));
    return e.result();
}

void sha256::encoder::reset() {
    SHA256_Init(&my.ctx);
}

void sha256::encoder::write(const char* d, uint32_t dlen) {
    SHA256_Update(&my.ctx,d,dlen);
}

sha256 sha256::encoder::result() {
    sha256 h;
    SHA256_Final((uint8_t*)h.data(),&my.ctx);
    return h;
}