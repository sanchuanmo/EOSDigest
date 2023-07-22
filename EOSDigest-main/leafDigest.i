%module(directors="1") eosdigest

%include <typemaps.i>
%include "std_string.i"
%include "std_vector.i"

%{
    #include "leafDigest.hpp"
%}

// %template(uint32_t) uint32_t
// %template(str) std::string
%template(StringVector) std::vector<std::string>;
%template(UCharVector) std::vector<unsigned char>;

extern std::string eosServializationTxDigest(const std::string status,const unsigned int cpu_usage_us,const unsigned int net_usage_words,const std::string compression,vector<unsigned char>& packed_trx,vector<std::string>& signatures, vector<unsigned char>& context_free_data);