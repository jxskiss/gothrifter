namespace go search

enum Corpus {
    UNIVERSAL = 0;
    WEB = 1;
    IMAGES = 2;
    LOCAL = 3;
    NEWS = 4;
    PRODUCTS = 5;
    VIDEO = 6;
}

struct Result {
    1: required string url;
    2: required string title;
    3: list<string> snippets;
    4: i64 post_at;
}

struct SearchRequest {
    1: string query;
    2: i32 page_number;
    3: i32 result_per_page;
    4: Corpus corpus;
}

struct SearchResponse {
    1: list<Result> results;
}

service SearchService {
    SearchResponse Search(1: SearchRequest req);
    void Ping();
    oneway void Ack(1: i64 some_id);
}
