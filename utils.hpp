// #pragma once
#include <iostream>
#include <iomanip>
#include<algorithm>
#include<cmath>
#include<cstdint>
#include<vector>
#include<cstring>
#include <sstream>
#include<assert.h>
#include<map>
using namespace std;



typedef unsigned char BYTE;

//字节数组转字符串
string bytesToString(vector<BYTE> &in){
    string out;
    out.assign(in.begin(),in.end());
    return out;
}

//字符串转字节数组
vector<BYTE> stringToBytes(string &in){
    vector<BYTE> out;
    out.assign(in.begin(),in.end());
    return out;
}

//十六进制字符串转字节数组
int hexToBytes(const string & in,unsigned char* out){
    int bytelen = in.length()/2;
    string strByte;
    unsigned int n;
    for(int i = 0; i < bytelen; ++i) {
        strByte = in.substr(i * 2,2);
        std::sscanf(strByte.c_str(),"%x",&n);
        *(out+i) = n;
    }

    return bytelen;
}



//字节数组转十六进制字符串
std::string bytestoHex(const char* d, uint32_t s) {
    std::string r;
    const char* to_hex = "0123456789abcdef";
    uint8_t* c = (uint8_t*)d;
    for (uint32_t i = 0; i < s; ++i)
        (r += to_hex[(c[i] >> 4)]) += to_hex[(c[i] & 0x0f)];
    return r;
}