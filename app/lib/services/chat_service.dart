import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/message.dart';

class ChatService {
  final String baseUrl;
  final http.Client client;

  ChatService({required this.baseUrl, http.Client? client})
      : client = client ?? http.Client();

  Future<String> sendMessages(List<Message> messages) async {
    final response = await client.post(
      Uri.parse('$baseUrl/chat'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'messages': messages.map((m) => m.toJson()).toList()}),
    );

    if (response.statusCode != 200) {
      throw Exception('No me pude conectar, inténtalo de nuevo');
    }
    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data['reply'] as String;
  }
}
